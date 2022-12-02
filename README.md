# About

This is a terraform provider for etcd.

Its scope is currently limited to the following resources:
- roles
- users
- keys
- range scoped states (to manage deletion of application states scoped by key ranges)
- synchronized key prefixes
- synchronized directory

We'll add further functionality as the need arises.

The provider can be found at: https://registry.terraform.io/providers/Ferlab-Ste-Justine/etcd/latest

# Local Troubleshoot

You need to have both golang 1.16 and Terraform setup on your machine for this to work. This also relies on a local minikube installation for running etcd.

## Setup Terraform to Detect Your Provider

Create a file named **.terraformrc** in your home directory with the following content:

```
provider_installation {
  dev_overrides {
    "ferlab/etcd" = "<Path to the project's parent directory on your machine>/terraform-provider-etcd"
  }
  direct {}
}
```

## Lauch the Etcd Server

Then, launch etcd by going to **test-environment/server** and typing:

```
terraform apply
```

Etcd will be exposed on port **32379**

Copy the **test-environment/server/certs** directory to **test-environment/provider/certs**.

## Build the Provider

Go to the root directory of this project and type:

```
go build .
```

## play with the Provider

From there, you can go to the **test-environment/provider** directory, edit the terraform scripts as you wish and experiment with the provider.

Note that you should not do **terraform init**. The provider was already setup globally in a previous step for your user.