package aws

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

//ServerTemplateManager defines Server template management functions a anyclouds provider must provide
type ServerTemplateManager struct {
	Provider *Provider
}

func toFloat32(s string) float32 {
	strFloat := strings.Replace(s, ",", "", -1)
	tmp, _ := strconv.ParseFloat(strFloat, 32)
	return float32(tmp)
}

func decodeMemory(mem string) int {
	tokens := strings.Split(mem, " ")
	if len(tokens) < 2 {
		return 0.
	}

	if tokens[1] == "GiB" {
		return int(toFloat32(tokens[0]) * 1024.)
	}
	return int(toFloat32(tokens[0]))
}

func decodeClockSpeed(cs string) float32 {
	tokens := strings.Split(cs, " ")
	if len(tokens) < 2 {
		return 0.
	}
	if tokens[1] == "Mhz" {
		return toFloat32(tokens[0]) / 1000.0
	}
	return toFloat32(tokens[0])
}

func decodeStorage(st string) int {
	tokens := strings.Split(st, " ")
	if len(tokens) < 3 {
		return 0.
	}
	nd, _ := strconv.Atoi(tokens[0])
	size, _ := strconv.Atoi(tokens[2])
	return nd * size
}

type Price struct {
	value float32
	date  time.Time
}

type PriceList []Price

// Forward request for length
func (p PriceList) Len() int {
	return len(p)
}

// Define compare
func (p PriceList) Less(i, j int) bool {
	return p[i].date.Before(p[j].date)
}

// Define swap over an array
func (p PriceList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func readPrice(js []byte) float32 {
	m := gjson.GetBytes(js, "terms.OnDemand").Map()

	plist := PriceList{}
	for _, res := range m {
		mp := res.Map()
		t, _ := time.Parse(time.RFC3339, mp["effectiveDate"].String())
		dim := mp["priceDimensions"]
		var value float32
		for _, res2 := range dim.Map() {
			value = float32(res2.Get("pricePerUnit.USD").Float())
		}
		p := Price{
			date:  t,
			value: value,
		}
		plist = append(plist, p)
	}
	if len(plist) == 0 {
		return float32(0)
	}
	sort.Sort(plist)
	return plist[0].value
}

func parseInt(s string) (int64, error) {
	i := 0
	for ; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' || s[i] == '+' || s[i] == '-' {
			break
		}
	}
	start := i
	if start >= len(s) {
		return 0, errors.Errorf("%s cannot be converted into int", s)
	}
	i++
	for ; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			continue
		} else {
			break
		}
	}
	stop := i

	if stop-start <= 0 {
		return 0, errors.Errorf("%s cannot be converted into int", s)
	}
	return strconv.ParseInt(s[start:stop], 10, 64)
}

func toTemplate(price aws.JSONValue) *api.ServerTemplate {
	js, err := json.Marshal(price)

	if err != nil {
		return nil
	}

	instanceName := gjson.GetBytes(js, "product.attributes.instanceType").String()
	creationDate, _ := time.Parse(time.RFC3339, gjson.GetBytes(js, "publicationDate").String())
	tpl := &api.ServerTemplate{
		ID:                instanceName, //
		Name:              instanceName,
		NumberOfCPUCore:   int(gjson.GetBytes(js, "product.attributes.vcpu").Int()),
		RAMSize:           decodeMemory(gjson.GetBytes(js, "product.attributes.memory").String()),
		SystemDiskSize:    20,
		EphemeralDiskSize: decodeStorage(gjson.GetBytes(js, "product.attributes.storage").String()),
		CreatedAt:         creationDate,
		Arch:              "",
		CPUFrequency:      decodeClockSpeed(gjson.GetBytes(js, "product.attributes.clockSpeed").String()),
		NetworkSpeed:      0,
		GPUInfo:           nil,
		OneDemandPrice:    readPrice(js),
	}

	instanceFamily := gjson.GetBytes(js, "product.attributes.instanceFamily").String()
	if instanceFamily == "GPU instance" {
		tpl.GPUInfo = &api.GPUInfo{
			Number:       int(gjson.GetBytes(js, "product.attributes.gpu").Int()),
			NumberOfCore: 0,
			MemorySize:   0,
			Type:         "",
		}
	}
	networkPerformance := gjson.GetBytes(js, "product.attributes.networkPerformance").String()
	speed, err := parseInt(networkPerformance)
	if err != nil {
		speed = 0
	} else {
		if strings.Contains(networkPerformance, "Gigabit") {
			speed = speed * 1000
		}
	}
	tpl.NetworkSpeed = int(speed)
	physicalProcessor := gjson.GetBytes(js, "product.attributes.physicalProcessor").String()
	processorArchitecture := gjson.GetBytes(js, "product.attributes.processorArchitecture").String()
	if strings.Contains(physicalProcessor, "Intel") ||
		strings.Contains(physicalProcessor, "AMD") &&
			processorArchitecture == "64-bit" {
		tpl.Arch = api.ArchAmd64
	} else if physicalProcessor == "Provider Graviton Processor" &&
		processorArchitecture == "64-bit" {
		tpl.Arch = api.ArchARM64
	} else {
		return nil
	}

	return tpl
}

func (mgr *ServerTemplateManager) createFilters() []*pricing.Filter {
	return []*pricing.Filter{
		{
			Field: aws.String("ServiceCode"),
			Type:  aws.String("TERM_MATCH"),
			Value: aws.String("AmazonEC2"),
		},
		{
			Field: aws.String("location"),
			Type:  aws.String("TERM_MATCH"),
			Value: aws.String(mgr.Provider.Configuration.RegionName),
		},
		{
			Type:  aws.String("TERM_MATCH"),
			Field: aws.String("capacitystatus"),
			Value: aws.String("Used"),
		},
		{
			Type:  aws.String("TERM_MATCH"),
			Field: aws.String("tenancy"),
			Value: aws.String("Shared"),
		},
		{
			Field: aws.String("preInstalledSw"),
			Type:  aws.String("TERM_MATCH"),
			Value: aws.String("NA"),
		},
		{
			Field: aws.String("operatingSystem"),
			Type:  aws.String("TERM_MATCH"),
			Value: aws.String("Linux"),
		},
		{
			Type:  aws.String("TERM_MATCH"),
			Field: aws.String("productFamily"),
			Value: aws.String("Compute Instance"),
		},
	}
}

//List returns available VM templates
func (mgr *ServerTemplateManager) List() ([]api.ServerTemplate, api.ListServerTemplatesError) {
	filters := mgr.createFilters()
	out, err := mgr.Provider.AWSServices.PricingClient.GetProducts(&pricing.GetProductsInput{
		Filters:       filters,
		MaxResults:    aws.Int64(100),
		FormatVersion: aws.String("aws_v1"),
		ServiceCode:   aws.String("AmazonEC2"),
	})
	if err != nil {
		return nil, api.NewListServerTemplatesError(err)
	}
	var result []api.ServerTemplate
	result = appendProducts(out, result)
	for err == nil && out != nil && out.NextToken != nil && out.PriceList != nil && len(out.PriceList) == 100 {
		out, err = mgr.Provider.AWSServices.PricingClient.GetProducts(&pricing.GetProductsInput{
			Filters:       filters,
			NextToken:     out.NextToken,
			FormatVersion: aws.String("aws_v1"),
			ServiceCode:   aws.String("AmazonEC2"),
			MaxResults:    aws.Int64(100),
		})
		result = appendProducts(out, result)
	}
	return result, api.NewListServerTemplatesError(err)
}

func appendProducts(out *pricing.GetProductsOutput, result []api.ServerTemplate) []api.ServerTemplate {
	for _, price := range out.PriceList {
		tpl := toTemplate(price)
		if tpl == nil || tpl.RAMSize == 0 {
			continue
		}
		result = append(result, *tpl)

	}
	return result
}

//Get returns the template identified by ids
func (mgr *ServerTemplateManager) Get(id string) (*api.ServerTemplate, api.GetServerTemplateError) {
	filters := append(mgr.createFilters(), &pricing.Filter{
		Field: aws.String("instanceType"),
		Type:  aws.String("TERM_MATCH"),
		Value: aws.String(id),
	})
	out, err := mgr.Provider.AWSServices.PricingClient.GetProducts(&pricing.GetProductsInput{
		Filters:       filters,
		FormatVersion: aws.String("aws_v1"),
		ServiceCode:   aws.String("AmazonEC2"),
	})
	if err != nil {
		return nil, api.NewGetServerTemplateError(err, id)
	}
	for _, price := range out.PriceList {
		res := toTemplate(price)
		if res.ID == id {
			return res, nil
		}

	}
	return nil, api.NewGetServerTemplateError(err, id)
}
