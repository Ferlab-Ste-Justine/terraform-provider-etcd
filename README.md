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

# Known Provider Quirk

At this point, if you provision users and roles with other resources, you might get the following provider error on etcd 3.5:

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x30 pc=0x98a870]

goroutine 37 [running]:
go.etcd.io/etcd/client/v3.(*Client).unaryClientInterceptor.func1({0xe2eff8?, 0xc00037c5a0?}, {0xd0c0f7, 0x1c}, {0xcce340, 0xc00037c600}, {0xcce4a0, 0xc0004d6320}, 0xc000034700, 0xd50580, ...)
	/home/eric/go/pkg/mod/go.etcd.io/etcd/client/v3@v3.5.5/retry_interceptor.go:81 +0x7f0
google.golang.org/grpc.(*ClientConn).Invoke(0xc00047b130?, {0xe2eff8?, 0xc00037c5a0?}, {0xd0c0f7?, 0xc00015a840?}, {0xcce340?, 0xc00037c600?}, {0xcce4a0?, 0xc0004d6320?}, {0x1362f40, ...})
	/home/eric/go/pkg/mod/google.golang.org/grpc@v1.41.0/call.go:35 +0x21f
...
```

If you apply multiple times, it will still go through.

Until the quirk is resolved, easiest way to circumvent this would be to provision users and roles separately from other resources.

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