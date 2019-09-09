package aws

import (
	"encoding/binary"
	"fmt"
	"github.com/SebastienDorgan/retry"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"sort"
	"time"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/pkg/errors"
)

//ServerManager defines Server management functions for AWS using EC2 interface
type ServerManager struct {
	AWS *Provider
}

func networkInterfaces(options *api.CreateServerOptions) []*ec2.InstanceNetworkInterfaceSpecification {
	var out []*ec2.InstanceNetworkInterfaceSpecification
	for i, sn := range options.Subnets {
		ni := ec2.InstanceNetworkInterfaceSpecification{
			AssociatePublicIpAddress:       nil,
			DeleteOnTermination:            nil,
			Description:                    nil,
			DeviceIndex:                    aws.Int64(int64(i)),
			Groups:                         nil,
			Ipv6AddressCount:               nil,
			Ipv6Addresses:                  nil,
			NetworkInterfaceId:             nil,
			PrivateIpAddress:               nil,
			PrivateIpAddresses:             nil,
			SecondaryPrivateIpAddressCount: nil,
			SubnetId:                       aws.String(sn),
		}
		out = append(out, &ni)
	}
	return out
}

func (mgr *ServerManager) createSpotInstance(options *api.CreateServerOptions, keyName string) (*string, error) {
	var blockDurationInMinutes *int64
	if options.SpotServerOptions.Duration > 0 {
		blockDuration := int64(options.SpotServerOptions.Duration / time.Minute)
		blockDurationInMinutes = aws.Int64(((blockDuration / 60) + 1) * 60)
	}

	input := &ec2.RequestSpotInstancesInput{
		InstanceCount: aws.Int64(1),
		LaunchSpecification: &ec2.RequestSpotLaunchSpecification{
			ImageId:      aws.String(options.ImageID),
			InstanceType: aws.String(options.TemplateID),
			KeyName:      aws.String(keyName),
			Placement: &ec2.SpotPlacement{
				AvailabilityZone: aws.String(mgr.AWS.AvailabilityZone),
			},
			NetworkInterfaces: networkInterfaces(options),
			UserData:          toString(options.BootstrapScript),
		},
		SpotPrice:            aws.String(fmt.Sprintf("%f", options.SpotServerOptions.HourlyPrice)),
		BlockDurationMinutes: blockDurationInMinutes,
		Type:                 aws.String("one-time"),
	}

	out, err := mgr.AWS.EC2Client.RequestSpotInstances(input)
	if err != nil || out.SpotInstanceRequests == nil {
		return nil, errors.Wrap(err, "Error creating spot instance")
	}

	req := out.SpotInstanceRequests[0]
	err = mgr.AWS.EC2Client.WaitUntilSpotInstanceRequestFulfilled(&ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []*string{req.SpotInstanceRequestId},
	})
	if err != nil {
		_, _ = mgr.AWS.EC2Client.CancelSpotInstanceRequests(&ec2.CancelSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []*string{req.SpotInstanceRequestId},
		})
		return nil, errors.Wrap(err, "error creating spot instance")
	}
	return mgr.getSpotInstanceId(req)

}

func (mgr *ServerManager) getSpotInstanceId(req *ec2.SpotInstanceRequest) (*string, error) {
	out2, err := mgr.AWS.EC2Client.DescribeSpotInstanceRequests(&ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []*string{req.SpotInstanceRequestId},
	})
	if err != nil {
		return nil, err
	}
	if len(out2.SpotInstanceRequests) == 0 {
		return nil, errors.Errorf("error getting spot request %s", *req.SpotInstanceRequestId)
	}
	return out2.SpotInstanceRequests[0].InstanceId, nil
}

func toString(reader io.Reader) *string {
	if reader == nil {
		return nil
	}
	b, _ := ioutil.ReadAll(reader)
	return aws.String(string(b))
}

func (mgr *ServerManager) createOnDemandInstance(options *api.CreateServerOptions, keyName string) (*string, error) {

	out, err := mgr.AWS.EC2Client.RunInstances(&ec2.RunInstancesInput{
		ImageId:           aws.String(options.ImageID),
		InstanceType:      aws.String(options.TemplateID),
		KeyName:           aws.String(keyName),
		NetworkInterfaces: networkInterfaces(options),
		UserData:          toString(options.BootstrapScript),
		Placement: &ec2.Placement{
			AvailabilityZone: aws.String(mgr.AWS.AvailabilityZone),
		},
		MinCount: aws.Int64(1),
		MaxCount: aws.Int64(1),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating on demand instance")
	}
	if out.Instances == nil || out.Instances[0].InstanceId == nil {
		return nil, errors.Errorf("unknow error creating reserved instance")
	}
	return out.Instances[0].InstanceId, nil
}

func (mgr *ServerManager) searchReservedInstanceOffering(options *api.CreateServerOptions) (*ec2.ReservedInstancesOffering, error) {
	tpl, err := mgr.AWS.TemplateManager.Get(options.TemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "Invalid template ID %s", options.TemplateID)
	}
	if tpl == nil {
		return nil, errors.Errorf("Unknow error getting server template %s", options.TemplateID)
	}
	out, err := mgr.AWS.EC2Client.DescribeReservedInstancesOfferings(&ec2.DescribeReservedInstancesOfferingsInput{
		AvailabilityZone:             aws.String(mgr.AWS.Region),
		DryRun:                       aws.Bool(false),
		Filters:                      nil,
		IncludeMarketplace:           aws.Bool(false),
		InstanceTenancy:              nil,
		InstanceType:                 aws.String(tpl.Name),
		MaxDuration:                  aws.Int64(int64(options.ReservedServerOptions.Duration / time.Second)),
		MaxInstanceCount:             aws.Int64(1),
		MaxResults:                   nil,
		MinDuration:                  nil,
		NextToken:                    nil,
		OfferingClass:                nil,
		OfferingType:                 nil,
		ProductDescription:           nil,
		ReservedInstancesOfferingIds: nil,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Error searching a reserved offering matching server options")
	}
	if out.ReservedInstancesOfferings == nil {
		return nil, nil
	}
	//sort by descending duration (max duration first)
	sort.Slice(out.ReservedInstancesOfferings, func(i, j int) bool {
		return *out.ReservedInstancesOfferings[i].Duration > *out.ReservedInstancesOfferings[j].Duration
	})
	return out.ReservedInstancesOfferings[0], nil
}

func (mgr *ServerManager) createReservedInstance(options *api.CreateServerOptions) (*string, error) {
	offering, err := mgr.searchReservedInstanceOffering(options)
	if err != nil {
		return nil, errors.Wrapf(err, "Error creating reserved instance")
	}
	if offering == nil {
		return nil, errors.Wrapf(
			errors.Errorf("No reserved instance offering matching server options found"),
			"Error creating reserved instance")
	}

	out, err := mgr.AWS.EC2Client.PurchaseReservedInstancesOffering(&ec2.PurchaseReservedInstancesOfferingInput{
		DryRun:                      aws.Bool(false),
		InstanceCount:               aws.Int64(1),
		LimitPrice:                  nil,
		ReservedInstancesOfferingId: offering.ReservedInstancesOfferingId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error creating reserved instance")
	}
	if out.ReservedInstancesId == nil {
		return nil, errors.Errorf("Unknow error creating reserved instance")
	}
	return out.ReservedInstancesId, nil
}

func (mgr *ServerManager) getInstance(instanceID string) retry.Action {
	return func() (v interface{}, e error) {
		return mgr.Get(instanceID)
	}
}

func (mgr *ServerManager) serverReady() retry.Condition {
	return func(v interface{}, e error) bool {
		if e != nil {
			return false
		}
		return v.(*api.Server).State == api.ServerReady
	}
}

func (mgr *ServerManager) associatePublicAddress(instanceID string) (string, error) {
	err := mgr.AWS.EC2Client.WaitUntilInstanceStatusOk(&ec2.DescribeInstanceStatusInput{
		InstanceIds: []*string{aws.String(instanceID)},
	})

	if err != nil {
		return "", err
	}
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
	var id *string
	var err error
	if options.SpotServerOptions != nil && options.ReservedServerOptions != nil {
		return nil, errors.Errorf("error creating server: SpotServerOptions and ReservedServerOptions are exclusive")
	}
	keyName := uuid.New().String()
	err = mgr.AWS.KeyPairManager.Import(keyName, options.KeyPair.PublicKey)
	defer func() { _ = mgr.AWS.KeyPairManager.Delete(keyName) }()
	if err != nil {
		return nil, errors.Wrapf(err, "error creating server %s", options.Name)
	}
	if options.SpotServerOptions != nil {
		id, err = mgr.createSpotInstance(options, keyName)
	} else if options.ReservedServerOptions != nil {
		id, err = mgr.createReservedInstance(options)
	} else {
		id, err = mgr.createOnDemandInstance(options, keyName)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error creating server %s", options.Name)
	}
	err = mgr.AWS.EC2Client.WaitUntilInstanceStatusOk(&ec2.DescribeInstanceStatusInput{
		InstanceIds: []*string{id},
	})

	if err != nil {
		_ = mgr.Delete(*id)
		return nil, errors.Wrapf(err, "error creating server %s", options.Name)
	}
	err = mgr.AWS.AddTags(*id, map[string]string{"name": options.Name})
	if err != nil {
		_ = mgr.Delete(*id)
		return nil, errors.Wrapf(err, "error creating server %s", options.Name)
	}
	err = mgr.addSecurityGroups(options, *id)
	if err != nil {
		_ = mgr.Delete(*id)
		return nil, errors.Wrapf(err, "error creating server %s", options.Name)
	}
	if options.PublicIP {
		_, err = mgr.associatePublicAddress(*id)
		if err != nil {
			_ = mgr.Delete(*id)
			return nil, errors.Wrapf(err, "error creating server %s", options.Name)
		}
	}
	err = mgr.AWS.EC2Client.WaitUntilInstanceStatusOk(&ec2.DescribeInstanceStatusInput{
		InstanceIds: []*string{id},
	})
	if err != nil {
		_ = mgr.Delete(*id)
		return nil, errors.Wrapf(err, "error creating server %s", options.Name)
	}
	return mgr.Get(*id)

}

func (mgr *ServerManager) addSecurityGroups(options *api.CreateServerOptions, instanceId string) error {
	if len(options.SecurityGroups) == 0 {
		return nil
	}
	var err error
	for _, sg := range options.SecurityGroups {
		_, err = mgr.AWS.EC2Client.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
			Groups:     []*string{aws.String(sg)},
			InstanceId: aws.String(instanceId),
		})
		if err != nil {
			return err
		}
	}
	err = mgr.AWS.EC2Client.WaitUntilInstanceStatusOk(&ec2.DescribeInstanceStatusInput{
		InstanceIds: []*string{&instanceId},
	})
	if err != nil {
		return err
	}
	return nil
}

//Delete delete Server identified by id
func (mgr *ServerManager) Delete(id string) error {
	srv, err := mgr.Get(id)
	releaseAddress := false
	if err == nil {
		//the server is associated to an elastic ip
		if len(srv.PublicIPv4) > 0 || len(srv.AccessIPv6) > 0 {
			releaseAddress = true
		}
	}
	_, err = mgr.AWS.EC2Client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	})

	if releaseAddress {
		err = mgr.releaseAddress(srv)
		if err != nil {
			return errors.Wrapf(err, "error deleting instance %s", id)
		}

	}
	err = mgr.AWS.EC2Client.WaitUntilInstanceTerminated(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	})
	return errors.Wrapf(err, "error deleting instance %s", id)
}

func (mgr *ServerManager) releaseAddress(srv *api.Server) error {
	out, err := mgr.AWS.EC2Client.DescribeAddresses(&ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("instance-id"),
				Values: []*string{aws.String(srv.ID)},
			},
		},
	})
	if err != nil {
		return errors.Wrapf(err, "error releasing elastic ip of instance %s", srv.ID)
	}
	if out == nil || out.Addresses == nil {
		return errors.Errorf("error releasing elastic ip of instance %s", srv.ID)
	}
	for _, addr := range out.Addresses {
		if *addr.PublicIp != srv.PublicIPv4 && *addr.PublicIp != srv.AccessIPv6 {
			continue
		}
		_, err := mgr.AWS.EC2Client.DisassociateAddress(&ec2.DisassociateAddressInput{
			AssociationId: addr.AssociationId,
		})
		_, err = mgr.AWS.EC2Client.ReleaseAddress(&ec2.ReleaseAddressInput{
			AllocationId: addr.AllocationId,
		})
		if err != nil {
			return errors.Wrapf(err, "error releasing elastic ip of instance %s", srv.ID)
		}
		return nil
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

func codeValue(code *int64) uint8 {
	if code == nil {
		return 255
	}
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(*code))
	return bs[0]
}

func state(in *ec2.InstanceState) api.ServerState {
	switch codeValue(in.Code) {
	case 0:
		return api.ServerPending
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
	name := ""
	n := sort.Search(len(instance.Tags), func(i int) bool {
		return *instance.Tags[i].Key == "name"
	})
	if n < len(instance.Tags) {
		name = *instance.Tags[n].Value
	}
	leasingType := api.LeasingTypeOnDemand
	if instance.SpotInstanceRequestId != nil {
		leasingType = api.LeasingTypeOnDemand
	}
	publicIPv4 := ""
	if instance.PublicIpAddress != nil {
		publicIPv4 = *instance.PublicIpAddress
	}
	return &api.Server{
		ID:             *instance.InstanceId,
		Name:           name,
		ImageID:        *instance.ImageId,
		TemplateID:     *instance.InstanceType,
		SecurityGroups: groupsIds(instance.SecurityGroups),
		PrivateIPs:     ips(instance.NetworkInterfaces),
		PublicIPv4:     publicIPv4,
		State:          state(instance.State),
		CreatedAt:      *instance.LaunchTime,
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
