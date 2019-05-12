package aws

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/pkg/errors"
)

//ServerManager defines Server management functions an anyclouds provider must provide
type ServerManager struct {
	AWS *Provider
}

func networkInterfaces(options *api.CreateServerOptions) []*ec2.InstanceNetworkInterfaceSpecification {
	var out []*ec2.InstanceNetworkInterfaceSpecification
	for _, sn := range options.Subnets {
		ni := ec2.InstanceNetworkInterfaceSpecification{
			SubnetId: aws.String(sn),
		}
		out = append(out, &ni)
	}
	return out
}

func (mgr *ServerManager) createSpotInstance(options *api.CreateServerOptions) (*api.Server, error) {
	tpl, err := mgr.AWS.GetTemplateManager().Get(options.TemplateID)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating spot instance")
	}
	input := &ec2.RequestSpotInstancesInput{
		InstanceCount: aws.Int64(1),
		LaunchSpecification: &ec2.RequestSpotLaunchSpecification{
			ImageId:      aws.String(options.ImageID),
			InstanceType: aws.String(options.TemplateID),
			KeyName:      aws.String(options.KeyPairName),
			Placement: &ec2.SpotPlacement{
				AvailabilityZone: aws.String(mgr.AWS.Region),
			},
			SecurityGroupIds:  aws.StringSlice(options.SecurityGroups),
			NetworkInterfaces: networkInterfaces(options),
			UserData:          aws.String(toString(options.BootstrapScript)),
		},
		SpotPrice: aws.String(fmt.Sprintf("%f", tpl.OneDemandPrice/4.0)),
		Type:      aws.String("one-time"),
	}

	out, err := mgr.AWS.EC2Client.RequestSpotInstances(input)
	if err != nil || out.SpotInstanceRequests == nil {
		return nil, errors.Wrap(err, "Error creating spot instance")
	}
	req := out.SpotInstanceRequests[0]
	if req.InstanceId == nil {
		return nil, errors.Wrap(err, "Error creating spot instance")
	}
	srv, err := mgr.Get(*req.InstanceId)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating spot instance")
	}
	return srv, nil

}

func toString(reader io.Reader) string {
	b, _ := ioutil.ReadAll(reader)
	return string(b)
}

func (mgr *ServerManager) createOnDemandInstance(options *api.CreateServerOptions) (*api.Server, error) {
	out, err := mgr.AWS.EC2Client.RunInstances(&ec2.RunInstancesInput{
		ImageId:           aws.String(options.ImageID),
		InstanceType:      aws.String(options.TemplateID),
		KeyName:           aws.String(options.KeyPairName),
		SecurityGroupIds:  aws.StringSlice(options.SecurityGroups),
		NetworkInterfaces: networkInterfaces(options),
		UserData:          aws.String(toString(options.BootstrapScript)),
		Placement: &ec2.Placement{
			AvailabilityZone: aws.String(mgr.AWS.Region),
		},
		MinCount: aws.Int64(1),
		MaxCount: aws.Int64(1),
	})
	if err != nil || out.Instances == nil {
		return nil, errors.Wrap(err, "Error creating on demand instance")
	}
	srv, err := mgr.Get(*out.Instances[0].InstanceId)
	if err != nil || out.Instances == nil {
		return nil, errors.Wrap(err, "Error creating on demand instance")
	}
	return srv, nil
}

func (mgr *ServerManager) createReservedInstance(options *api.CreateServerOptions) (*api.Server, error) {
	//out, err := mgr.AWS.EC2Client.PurchaseReservedInstancesOffering(&ec2.PurchaseReservedInstancesOfferingInput{
	//
	//})
	return nil, nil
}

func (mgr *ServerManager) associatePublicAddress(instanceID string) (string, error) {
	allocRes, err := mgr.AWS.EC2Client.AllocateAddress(&ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
	})
	if err != nil {
		return "", err
	}
	_, err = mgr.AWS.EC2Client.AssociateAddress(&ec2.AssociateAddressInput{
		AllocationId: allocRes.AllocationId,
		InstanceId:   aws.String(instanceID),
	})
	if err != nil {
		return "", err
	}
	return *allocRes.PublicIp, err
}

//Create creates an Server with options
func (mgr *ServerManager) Create(options *api.CreateServerOptions) (*api.Server, error) {
	var srv *api.Server
	var err error
	if options.LeasingType == api.LeasingTypeSpot {
		srv, err = mgr.createSpotInstance(options)
	} else if options.LeasingType == api.LeasingTypeReserved {
		srv, err = mgr.createReservedInstance(options)
	} else {
		srv, err = mgr.createOnDemandInstance(options)
	}
	if err != nil {
		return srv, errors.Wrap(err, "error creating server")
	}
	if !options.PublicIP {
		return srv, nil
	}
	_, err = mgr.associatePublicAddress(srv.ID)
	if err != nil {
		return srv, errors.Wrapf(err, "error associating elastic ip to server %s", srv.ID)
	}
	return mgr.Get(srv.ID)
}

//Delete delete Server identified by id
func (mgr *ServerManager) Delete(id string) error {
	_, err := mgr.AWS.EC2Client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	})
	if err != nil {
		return errors.Wrapf(err, "Error deleting instance %s", id)
	}
	return nil
}

func (mgr *ServerManager) getReservedInstances() ([]*ec2.ReservedInstances, error) {
	out, err := mgr.AWS.EC2Client.DescribeReservedInstances(&ec2.DescribeReservedInstancesInput{})
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return out.ReservedInstances, nil
}

func mapToSlice(in map[string]*api.Server) []api.Server {
	out := make([]api.Server, len(in))
	for _, v := range in {
		out = append(out, *v)
	}
	return out
}

//List list Servers
func (mgr *ServerManager) List() ([]api.Server, error) {
	out, err := mgr.AWS.EC2Client.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, errors.Wrap(err, "error listing instances")
	}
	serverMap := map[string]*api.Server{}
	for _, res := range out.Reservations {
		for _, ins := range res.Instances {
			srv := server(ins)
			serverMap[srv.ID] = srv
		}
	}
	ris, err := mgr.getReservedInstances()
	if err != nil { //No reserved instance
		return mapToSlice(serverMap), nil
	}
	for _, ri := range ris {
		srv, ok := serverMap[*ri.ReservedInstancesId]
		if !ok {
			continue
		}
		srv.LeasingType = api.LeasingTypeReserved
		srv.CreatedAt = *ri.Start
		srv.LeaseDuration = ri.End.Sub(*ri.Start)
		srv.LeasingType = api.LeasingTypeReserved
	}

	return mapToSlice(serverMap), nil
}

func ips(in []*ec2.InstanceNetworkInterface) map[api.IPVersion][]string {
	ipMap := make(map[api.IPVersion][]string)
	ipMap[api.IPVersion4] = []string{}
	ipMap[api.IPVersion6] = []string{}
	for _, netIF := range in {
		ipMap[api.IPVersion4] = append(ipMap[api.IPVersion4], *netIF.PrivateIpAddress)
		for _, v6ip := range netIF.Ipv6Addresses {
			ipMap[api.IPVersion6] = append(ipMap[api.IPVersion6], *v6ip.Ipv6Address)
		}
	}
	return ipMap
}

func state(in *ec2.InstanceState) api.ServerState {
	switch *in.Code {
	case 0:
		return api.ServerBuilding
	case 16:
		return api.ServerReady
	case 32:
		return api.ServerShutoff
	case 48:
		return api.ServerDeleted
	case 64:
		return api.ServerShutoff
	case 80:
		return api.ServerShutoff
	default:
		return api.ServerInError
	}
}
func groupsIds(in []*ec2.GroupIdentifier) []string {
	var result []string
	for _, g := range in {
		result = append(result, *g.GroupId)
	}
	return result
}
func server(instance *ec2.Instance) *api.Server {
	leasingType := api.LeasingTypeOnDemand
	if instance.SpotInstanceRequestId != nil {
		leasingType = api.LeasingTypeOnDemand
	}
	return &api.Server{
		ID:             *instance.InstanceId,
		Name:           *instance.PrivateDnsName,
		ImageID:        *instance.ImageId,
		TemplateID:     *instance.InstanceType,
		SecurityGroups: groupsIds(instance.SecurityGroups),
		PrivateIPs:     ips(instance.NetworkInterfaces),
		PublicIPv4:     *instance.PublicIpAddress,
		State:          state(instance.State),
		CreatedAt:      *instance.LaunchTime,
		KeyPairName:    *instance.KeyName,
		LeasingType:    leasingType,
	}
}

func (mgr *ServerManager) getReservedInstance(id string) (*ec2.ReservedInstances, error) {
	out, err := mgr.AWS.EC2Client.DescribeReservedInstances(&ec2.DescribeReservedInstancesInput{
		ReservedInstancesIds: []*string{aws.String(id)},
	})
	if err != nil {
		return nil, err
	}
	if out == nil || out.ReservedInstances == nil {
		return nil, errors.Errorf("Instance %s is not a reserved instance", id)
	}
	return out.ReservedInstances[0], nil
}

//Get get Servers
func (mgr *ServerManager) Get(id string) (*api.Server, error) {

	out, err := mgr.AWS.EC2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(id),
		},
	})
	if err != nil || out.Reservations == nil || out.Reservations[0].Instances == nil {
		return nil, errors.Wrapf(err, "error getting instance %s", id)
	}
	srv := server(out.Reservations[0].Instances[0])
	if srv.LeasingType == api.LeasingTypeSpot {
		return srv, nil
	}
	ri, err := mgr.getReservedInstance(id)
	if err != nil { //On demand instance
		return srv, nil
	}
	srv.CreatedAt = *ri.Start
	srv.LeaseDuration = ri.End.Sub(*ri.Start)
	srv.LeasingType = api.LeasingTypeReserved

	return srv, nil
}

//Start starts an Server
func (mgr *ServerManager) Start(id string) error {
	_, err := mgr.AWS.EC2Client.StartInstances(&ec2.StartInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	})
	if err != nil {
		return errors.Wrapf(err, "Error starting instance %s", id)
	}
	return nil
}

//Stop stops an Server
func (mgr *ServerManager) Stop(id string) error {
	_, err := mgr.AWS.EC2Client.StopInstances(&ec2.StopInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	})
	if err != nil {
		return errors.Wrapf(err, "Error starting instance %s", id)
	}
	return nil
}

//Resize resize a server
func (mgr *ServerManager) Resize(id string, templateID string) error {
	_, err := mgr.AWS.OpsWorksClient.UpdateInstance(&opsworks.UpdateInstanceInput{})
	if err != nil {
		return errors.Wrapf(err, "Error resizing instance %s", id)
	}
	return nil
}
