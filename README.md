# Varus
Simple cli helper tool written in go to get latest AWS EKS AMI release version and compare it with your kubernetes cluster release version

### Install
```
wget https://github.com/bhegazy/varus/releases/download/0.0.1/varus-darwin-amd64
chmod +x varus-darwin-amd64
mv /usr/local/bin/varus
```


### Example Usage

```
> export AWS_ACCESS_KEY_ID=xxxxxxx
export AWS_SECRET_ACCESS_KEY=xxx
export AWS_REGION=xxxx

> varus get -k 1.18 # Get the latest EKS AMI Release Version
Latest EKS ami release version: 1.18.9-20210208 # Output

❯ varus compare -k 1.18
+---+--------------------+-----------------------+-------------------------+------------------------+---------------+
|   | CLUSTER NAME       | NODEGROUP NAME        | CURRENT RELEASE VERSION | LATEST RELEASE VERSION | USING LATEST? |
+---+--------------------+-----------------------+-------------------------+------------------------+---------------+
| 1 | k8s-example        | k8s-example-nodegroup | 1.18.9-20210125         | 1.18.9-20210208        | No ⚔️         |
+---+--------------------+-----------------------+-------------------------+------------------------+---------------+
```
### the cli can be used with [aws-vault](https://github.com/99designs/aws-vault) without exporting AWS creds

> This is useful when u have multiple aws accounts

```
> aws-vault exec <ur-aws-account-profile> -- varus compare -k 1.18
+---+---------------+----------------------+-------------------------+------------------------+---------------+
|   | CLUSTER NAME  | NODEGROUP NAME       | CURRENT RELEASE VERSION | LATEST RELEASE VERSION | USING LATEST? |
+---+---------------+----------------------+-------------------------+------------------------+---------------+
| 1 | cluster1      | cluster1-nodegroup-1 | 1.18.9-20201117         | 1.18.9-20210208        | No ⚔️         |
| 2 | cluster2      | cluster2-nodegroup-1 | 1.18.9-20210125         | 1.18.9-20210208        | No ⚔️         |
| 3 | cluster3      | cluster3-nodegroup-1 | 1.18.8-20201007         | 1.18.9-20210208        | No ⚔️         |
| 4 | cluster3      | cluster3-nodegroup-2 | 1.18.9-20210112         | 1.18.9-20210208        | No ⚔️         |
| 5 | cluster3      | cluster3-nodegroup-3 | 1.18.9-20201117         | 1.18.9-20210208        | No ⚔️         |
+---+---------------+----------------------+-------------------------+------------------------+---------------+
```
