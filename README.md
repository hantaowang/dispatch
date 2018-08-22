# dispatch
Allocate and manage scoped Kubernetes namespaces.

## About
**Dispatch** is a tool allows cluster admins to generate and hand out namespace
permissions to users.

If a developer wants a namespace to run their application, **dispatch** can create
the namespace and also a scoped user that is only able to edit and view resources
in that particular namespace.

It also allows a single user to own multiple namespaces, or for multiple users to own
the same namespace.

**Dispatch** is currently a proof of concept tool. While it does the above functions,
it is not ready to be used in a live environment.

## How it Works
**Dispatch** manages namespaces and users using two Custom Resource Definitions. 
A `DispatchUser` is a developer or user who wishes to own namespaces. It requires a unique
identifier and a list of namespaces it wants to own. 
A `OwnedNamespace` (will be changed to `NamespaceBinding` in the future) is resource that
signifies that a user should have access to a namespace.

**Dispatch** has two controllers. One takes a `DispatchUser` and generates or deletes 
`OwnedNamespace` objects depending on what namespaces a user wishes to use. It also
creates a `ServiceAccount` for each user. A second controller watches `OwnedNamespace` objects
and creates a `RoleBinding` that binds the default `edit` `ClusterRole` to the user's `ServiceAccount`,
scoped to only have permissions in the namespace specified.

## How to Use

Currently **Dispatch** can only be used by creating and modifying `DispatchUser` objects through
`kubectl`. 

    kubectl apply -f manifests/crd.yaml    # Set up cluster, create CRDs
    go run main.go                         # Run dispatch

To try it out, run `kubectl apply -f manifests/testdispatchuser.yaml`. A service account `123456` 
should be created in the `dispatch` namespace along with a secret `123456-token-*`. Then `token` field
of that secret is the token to authenticate with. This can be done by creating a User in `~/.kube/config`
with that token and using that User. This User wll only be able to edit and view `test-namespace-2`.

### Future Plans (TODO)
Unless there is real interest in this project, these future plans will not be implemented since they are
not really that interesting. These are just some features that would need to exist if this tool were to be
used in a non experimental environment. This project currently exists just as something fun to do while I 
was bored one weekend. But none the less, here are some work that I think would be useful.  

- Create a Python server that authenticates users with Google OAuth and allows users to request namespaces
- Have this server automatically create and modify the `DispatchUser` objects.
- Automatically set up the authentication in `~/.kube/config`.
- Use a custom `ClusterRole` rather than the default `edit` to scope permissions further.
- In the Python server validate that the User has permission to access the requested namespace.
- The ability to enforce resource limits on namespaces.
- Integration with GKE or EKS that allows auto scaling of the cluster based on usage.

