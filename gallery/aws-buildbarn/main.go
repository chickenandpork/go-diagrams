package main

import (
	"log"

	"github.com/blushft/go-diagrams/diagram"
	"github.com/blushft/go-diagrams/nodes/aws"
	"github.com/blushft/go-diagrams/nodes/k8s"
)

func main() {

	d, err := diagram.New(diagram.Filename("buildbarn"), diagram.Label("Build Cluster"), diagram.Direction("TB"))
	if err != nil {
		log.Fatal(err)
	}

	users := aws.General.Users(diagram.NodeLabel("Users (web)"))
	bazel := aws.General.Users(diagram.NodeLabel("Users (bazel)"))
	lb_stor := k8s.Network.Ing(diagram.NodeLabel("Ingress (grpc)"))
	lb_bep := k8s.Network.Ing(diagram.NodeLabel("Ingress (BEP)"))
	lb := k8s.Network.Ing(diagram.NodeLabel("Ingress (web)"))

	dc := diagram.NewGroup("GCP")
	dc.NewGroup("ingress").
		Label("Ingress Layer").
		Add(
			lb, lb_stor, lb_bep,
		)

	sched := k8s.Compute.Pod(diagram.NodeLabel("bb_scheduler"))
	portal := k8s.Compute.Pod(diagram.NodeLabel("bb_portal"))
	bs1 := k8s.Compute.Pod(diagram.NodeLabel("bb_Storage 1"))
	bs2 := k8s.Compute.Pod(diagram.NodeLabel("bb_Storage 2"))
	bs3 := k8s.Compute.Pod(diagram.NodeLabel("bb_Storage 3"))
	dc.NewGroup("bb_storage").
		Label("Storage (& Routing) Layer").
		Add( bs1, bs2, bs3).
		ConnectAllFrom(lb_stor.ID(), diagram.Forward()).
		ConnectAllTo(sched.ID(), diagram.Forward()).
		Add(sched)

	dc.NewGroup("bb_workers").
		Label("Workers/Runners").
		Add(
			k8s.Compute.Pod(diagram.NodeLabel("Worker 1")),
			k8s.Compute.Pod(diagram.NodeLabel("Worker 2")),
			k8s.Compute.Pod(diagram.NodeLabel("Worker 3")),
		).
		ConnectAllFrom(sched.ID(), diagram.Reverse()).
		ConnectAllFrom(bs1.ID(), diagram.Reverse())


	dc.NewGroup("services").
		Label("Service Layer").
		Add(
			k8s.Compute.Pod(diagram.NodeLabel("bb_browser")),
			portal,
		).
		ConnectAllFrom(lb.ID(), diagram.Forward())

	d.Connect(users, lb, diagram.Forward()).Group(dc)
	d.Connect(bazel, lb_stor, diagram.Forward()).Group(dc)
	d.Connect(bazel, lb_bep, diagram.Forward()).Group(dc)
	d.Connect(lb_bep, portal, diagram.Forward()).Group(dc)
	d.Connect(portal, sched, diagram.Forward()).Group(dc)

	if err := d.Render(); err != nil {
		log.Fatal(err)
	}
}
