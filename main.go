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
var (
	errCodeInvalidK8sVersion = "Invalid k8s version provided..."
)

func main() {
	newSession := newSession()
	awsConfig := aws.NewConfig()

	getCommand := flag.NewFlagSet("get", flag.ExitOnError)
	getK8sVersion := getCommand.String("k", "", "AWS EKS kubernetes version (Required)")

	compareCommand := flag.NewFlagSet("compare", flag.ExitOnError)
	compareK8sVersion := compareCommand.String("k", "", "AWS EKS kubernetes version (Required)")

	if len(os.Args) < 2 {
		fmt.Println("get or compare subcommand is required")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "get":
		getCommand.Parse(os.Args[2:])
	case "compare":
		compareCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
	if getCommand.Parsed() {
		// Required Flags
		if *getK8sVersion == "" {
			getCommand.PrintDefaults()
			os.Exit(1)
		}
		latestVersion, err := getLatestVersion(newSession, releaseParam(*getK8sVersion))
		if err != nil {
			awsErr(err)
		} else {
			fmt.Printf("Latest EKS ami release version: %s\n",
				latestVersion,
			)
		}
	}
	if compareCommand.Parsed() {
		// Required Flags
		if *compareK8sVersion == "" {
			compareCommand.PrintDefaults()
			os.Exit(1)
		}
		latestVersion, err := getLatestVersion(newSession, releaseParam(*compareK8sVersion))
		awsErr(err)
		svc := eks.New(newSession, awsConfig)
		clusterList, err := listClusters(svc)
		awsErr(err)
		t := tableInit()
		for _, c := range clusterList {
			nodeGroupName, err := getNodegroupName(svc, *c)
			awsErr(err)
			for _, n := range nodeGroupName {
				nodegroupVersion, err := getNodegroupVersion(svc, *c, *n)
				awsErr(err)
				t.AppendRow(table.Row{ *c, *n, nodegroupVersion, latestVersion})
			}
		}
		if len(clusterList) == 0 {
			fmt.Println("No EKS clusters found in this AWS account or region...")
		} else {
			fmt.Println(t.Render())
		}
	}
}

func tableInit() table.Writer {
	t := table.NewWriter()
	t.SetAutoIndex(true)
	t.AppendHeader(table.Row{"Cluster Name", "Nodegroup Name", "Current Release Version", "Latest Release Version" })
	return t
}

func newSession() *session.Session {
	s := session.Must(session.NewSession())
	return s
}

func listClusters(svc *eks.EKS)  ([]*string, error) {
	listClustersInput := &eks.ListClustersInput{}
	listClusters, err := svc.ListClusters(listClustersInput)
	if err != nil {
		return nil, err
	}
	return listClusters.Clusters, nil
}

func releaseParam(v string) string {
	return "/aws/service/eks/optimized-ami/" + v + "/amazon-linux-2-arm64/recommended/release_version"
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

func awsErr(err error) {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeParameterNotFound:
				fmt.Println(errCodeInvalidK8sVersion, aerr.Error())
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
