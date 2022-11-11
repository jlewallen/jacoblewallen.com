+++
title = "LXC Is Awesome"
date = 2011-05-15T00:00:00-08:00
draft = false
authors = ["jacob"]
[taxonomies]
tags = ["linux"]
+++

I don’t think I can stress how excited I was after discovering Linux Containers. With all of the attention being given to cloud computing and virtualization I’m surprised more people haven’t been talking about LXC. If you haven’t heard of them, here’s a good description from the docs:

> Linux Containers take a completely different approach than system virtualization technologies such as KVM and Xen, which started by booting separate virtual systems on emulated hardware and then attempted to lower their overhead via paravirtualization and related mechanisms. Instead of retrofitting efficiency onto full isolation, LXC started out with an efficient mechanism (existing Linux process management) and added isolation, resulting in a system virtualization mechanism as scalable and portable as chroot, capable of simultaneously supporting thousands of emulated systems on a single server while also providing lightweight virtualization options to routers and smart phones.

If you’re convinced of the benefits of virtualization, you’re probably already interested. Many of the benefits of virtualization apply to linux containers, as well as a few others:

# Fast and Easy

Once you’ve done the initial footwork, setting up new containers is very simple and as easy as tweaking some configuration and uncompressing the archive containing the container’s root filesystem. Once a container is running, all of the processes in the container are visible from the host machine, so tools like ps and htop will show you everything that’s going on.

# Resource Efficient

Unlike virtual machines, containers don’t need to be given fixed amounts of memory or CPU. Unless you’re taking advantage of the overcommitting features available in most virtualization technologies you’re fairly confined on how you can divide resources among virtual machines. With linux containers you have far more control and flexibility. Simply create containers and allow the operating system to divide resources appropriately. You can also use process limits to gain any extra control you may want.

# Useful

I’ve found several uses recently for LXC and I’ll be posting more about them in the future, here are just a few:

1. A platform for developing Chef/Puppet scripts.
2. Throw away servers for experimentation.
3. Configuring and testing cluster configurations.

As a work-from-home software engineer with a single server machine beside my desk it’s made my life much more comfortable.

# Links

* [State of LXC in Ubuntu 11.04](http://www.stgraber.org/2011/05/04/state-of-lxc-in-ubuntu-natty/)
* [LXC Linux Containers](http://lxc.sourceforge.net/)
* [LXC Configure Ubuntu Lucid Containers](http://blog.bodhizazen.net/linux/lxc-configure-ubuntu-lucid-containers/)
