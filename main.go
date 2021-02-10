package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/jedib0t/go-pretty/table"
	"os"
)

func newSession() *session.Session {
	s:= session.Must(session.NewSession())
	return s
}

func main() {
	k8sVersion := flags()
	releaseParamName := "/aws/service/eks/optimized-ami/" + k8sVersion + "/amazon-linux-2-arm64/recommended/release_version"
	newSession := newSession()

	latestVersion, err := getLatestVersion(newSession, releaseParamName)
	if err != nil {
		panic(err)
	}

	awsConfig := aws.NewConfig()
	svc := eks.New(newSession, awsConfig)

	clusterList, err := listClusters(svc)
	eksErr(err)
	t := tableInit()
	for _, c := range clusterList {
		nodeGroupName, err := getNodegroupName(svc, *c)
		eksErr(err)
		for _, n := range nodeGroupName {
			nodegroupVersion, err := getNodegroupVersion(svc, *c, *n)
			eksErr(err)
			t.AppendRow(table.Row{ *c, *n, nodegroupVersion, latestVersion})
		}
	}
	fmt.Println(t.Render())
}

func listClusters(svc *eks.EKS)  ([]*string, error ) {
	listClustersInput := &eks.ListClustersInput{}
	listClusters, err := svc.ListClusters(listClustersInput)
	if err != nil {
		return  nil, err
	}
	return listClusters.Clusters, nil
}

func tableInit() table.Writer {
	t := table.NewWriter()
	t.SetAutoIndex(true)
	t.AppendHeader(table.Row{"Cluster Name", "Nodegroup Name", "Current Release Version", "Latest Release Version" })
	return t
}

func flags() string {
	// Command args: check-latest, compare
	// check-latest will check for latest ami release version
	// compare will compare all current eks clusters/nodegroup with latest release version
	// both commands will require flag --k8s-version
	checkCommand := flag.NewFlagSet("check-latest", flag.ExitOnError)
	compareCommand := flag.NewFlagSet("compare", flag.ExitOnError)

	checkK8sVersion := checkCommand.String("k8s-version", "", "AWS EKS kubernetes version (Required)")
	compareK8sVersion := compareCommand.String("k8s-version", "", "AWS EKS kubernetes version (Required)")

	if len(os.Args) < 2 {
		fmt.Println("check-latest or compare subcommand is required")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "check-latest":
		checkCommand.Parse(os.Args[2:])
	case "compare":
		compareCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
	if checkCommand.Parsed() {
		// Required Flags
		if *checkK8sVersion == "" {
			checkCommand.PrintDefaults()
			os.Exit(1)
		}
		// Print
		fmt.Printf("checkK8sVersion: %s\n",
			*checkK8sVersion,
		)
		return *checkK8sVersion
	}
	if compareCommand.Parsed() {
		// Required Flags
		if *compareK8sVersion == "" {
			compareCommand.PrintDefaults()
			os.Exit(1)
		}
		// Print
		fmt.Printf("compareK8sVersion: %s\n",
			*compareK8sVersion,
		)
		return *compareK8sVersion
	}
	return ""
}

func getLatestVersion(session *session.Session, name string) (string, error) {
	awsConfig := aws.NewConfig()
	svc := ssm.New(session, awsConfig)
	paramIn := ssm.GetParameterInput{
		Name:aws.String(name),
	}
	latestversion, err := svc.GetParameter(&paramIn)
	if err != nil {
		return "", err
	}
	return *latestversion.Parameter.Value, nil
}

func getNodegroupName(svc *eks.EKS, c string) ([]*string, error ){
	listNodegroupsInput := eks.ListNodegroupsInput{
		ClusterName: aws.String(c),
	}
	listNodegroups, err := svc.ListNodegroups(&listNodegroupsInput)
	if err != nil {
		return nil, err
	}
	return listNodegroups.Nodegroups, nil
}

func getNodegroupVersion(svc *eks.EKS, c, ng string) (string, error ){
	describeNodeGroupInput := eks.DescribeNodegroupInput{
		ClusterName: aws.String(c),
		NodegroupName: aws.String(ng),
	}
	describeNodeGroup, err := svc.DescribeNodegroup(&describeNodeGroupInput)
	if err != nil {
		return "", err
	}
	return *describeNodeGroup.Nodegroup.ReleaseVersion, nil
}

func eksErr(err error) {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case eks.ErrCodeResourceNotFoundException:
				fmt.Println(eks.ErrCodeResourceNotFoundException, aerr.Error())
			case eks.ErrCodeClientException:
				fmt.Println(eks.ErrCodeClientException, aerr.Error())
			case eks.ErrCodeServerException:
				fmt.Println(eks.ErrCodeServerException, aerr.Error())
			case eks.ErrCodeServiceUnavailableException:
				fmt.Println(eks.ErrCodeServiceUnavailableException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}
}
