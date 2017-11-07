# terraform-provisioner-puppet [experimental]

[![Build Status](https://travis-ci.org/pyToshka/terraform-provisioner-puppet.svg?branch=master)](https://travis-ci.org/pyToshka/terraform-provisioner-puppet)

> Provision terraform resources with puppet

## Overview

**[Terraform](https://github.com/hashicorp/terraform)** is a tool for automating infrastructure. Terraform includes the ability to provision resources at creation time through a plugin api. Currently, some builtin [provisioners](https://www.terraform.io/docs/provisioners/) such as **chef** and standard scripts are provided; this provisioner introduces the ability to provision an instance at creation time with **puppet**.

This provisioner provides the ability to install puppet agent and try to configure instance.

**terraform-provisioner-puppet** is shipped as a **Terraform** [module](https://www.terraform.io/docs/modules/create.html). To include it, simply download the binary and enable it as a terraform module in your **terraformrc**.

## Installation
```bash
go get github.com/pyToshka/terraform-provisioner-puppet

```
**terraform-provisioner-puppet** ships as a single binary and is compatible with **terraform**'s plugin interface. Behind the scenes, terraform plugins use https://github.com/hashicorp/go-plugin and communicate with the parent terraform process via RPC.

Once installed, a `~/.terraformrc` file is used to _enable_ the plugin.

```bash
providers {
    puppet = "/usr/local/bin/terraform-provisioner-puppet"
}
```

## Usage

Once installed, you can provision resources by including an `puppet` provisioner block.

The following example demonstrates a configuration block to install and running puppet agent to new instances.


```
resource "aws_instance" "puppetagent" {
    ami = "${var.ami_id}"
    instance_type = "${var.instance_type}"
    iam_instance_profile = "${var.iam_instance_profile}"
    count = 1
    tags {
        Name = "puppet-agent.example.com"
        puppet_role = "base"
        CNAME = "puppet-agent.example.com."
    }
    key_name = "${var.aws_key_name}"
    subnet_id = "${var.subnet_id}"
    security_groups = ["${var.security_group}"]
    vpc_security_group_ids = ["${var.security_group}"]
   provisioner "puppet" {
     connection {
       user = "ubuntu" //username for ssh connection
       private_key = "${file("key.pem")}"
     }
     puppetmaster_ip="172.22.61.191" #ip of Puppet Master
     use_sudo = true
   }
}
```

e

I'm not testing the custom_command. Full working are granted only with Ubuntu based distribution
================================================================================================
