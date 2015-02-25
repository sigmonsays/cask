container and image management tool

description
--------------
the initial goal of cask is to make it easy to manage containers and images

currently it supports building a image in a lxc against a given runtime. A
runtime is thought of as a base image your application will run against.

This is slightly different than a base OS image since its main purpose is
for deploying your application on a pre-defined application stack.  Such
examples would be a PHP application where all its dependencies have already
been included in the runtime.

runtime
--------------

Any image can be used as a runtime but here is a quick way to bootstrap a runtime with lxc

    lxc-create --template ubuntu --name ubuntu12  -- --release precise

quickstart
--------------
- requires lxc 1.0 
- ensure you have a runtime available

build your first container from this example directory

    sudo cask build -caskpath examples/golang/cask


