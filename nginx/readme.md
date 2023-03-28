# Reverse Proxy #

A reverse proxy is a server that sits in front of one or more backend servers and forwards client requests to the appropriate backend server. A reverse proxy provides an additional level of abstraction and control to ensure the smooth flow of network traffic between clients and servers. A reverse proxy may reside on the same server as one or more backend servers or may reside on a separate server that is networked to the backend servers. 

We will be using nginx as our reverse proxy.

# Load Balancing #

Load balancing is the process of distributing network or application traffic across a group of servers. Load balancing improves the overall performance of applications and websites by ensuring that no single server is overloaded while others are underutilized. Load balancing is also used to increase fault tolerance and availability of applications and websites. 

We will be using nginx as our load balancer.

# Encrypting Traffic #

SSL/TLS is a cryptographic protocol that provides communications security over a computer network. SSL/TLS works by using a public key to encrypt data before it is transmitted across the network, and a private key to decrypt the data once it has been received. SSL/TLS is used to secure credit card transactions, data transfers and logins, and more recently is becoming the norm when securing browsing of websites. TLS (Transport Layer Security) is just an updated, more secure, version of SSL.

We will be using Let's Encrypt to generate our SSL/TLS certificates.

# Let's Encrypt #

Let’s Encrypt is a Certificate Authority (CA) that provides an easy way to obtain and install free TLS/SSL certificates, thereby enabling encrypted HTTPS on web servers. It simplifies the process by providing a software client, Certbot, that attempts to automate most (if not all) of the required steps. Let’s Encrypt is a service provided by the Internet Security Research Group (ISRG).

# Kubernetes #

Coming soon

# Ingress #

Coming soon
<!-- # Kubernetes #

Kubernetes is an open-source system for automating deployment, scaling, and management of containerized applications. It groups containers that make up an application into logical units for easy management and discovery. Kubernetes builds upon 15 years of experience of running production workloads at Google, combined with best-of-breed ideas and practices from the community.

# Helm #

Helm is a package manager for Kubernetes. It allows you to define, install, and upgrade even the most complex Kubernetes application. Helm charts help you define, install, and upgrade even the most complex Kubernetes application.

# Ingress #

In Kubernetes, an Ingress is an object that allows access from outside the cluster to services running inside the cluster. Ingress can provide load balancing, SSL termination and name-based virtual hosting. Ingress may provide other features, depending on the ingress controller being used.

# NGINX Ingress Controller #

The NGINX Ingress Controller is an Ingress controller for Kubernetes using NGINX as a reverse proxy and load balancer. The Ingress controller is deployed as a Kubernetes Pod. The Ingress resource configures the load balancing rules and defines the services to proxy traffic to. -->
