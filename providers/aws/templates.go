package aws

import (
	"encoding/json"
	"fmt"
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
	AWS *Provider
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
		return toFloat32(tokens[0]) * 1000.0
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

func toTemplate(price aws.JSONValue) *api.ServerTemplate {
	js, err := json.Marshal(price)

	if err != nil {
		return nil
	}
	creationDate, _ := time.Parse(time.RFC3339, gjson.GetBytes(js, "publicationDate").String())
	tpl := &api.ServerTemplate{
		ID:                gjson.GetBytes(js, "product.sku").String(),
		Name:              gjson.GetBytes(js, "product.attributes.instanceType").String(),
		NumberOfCPUCore:   int(gjson.GetBytes(js, "product.attributes.vcpu").Int()),
		RAMSize:           decodeMemory(gjson.GetBytes(js, "product.attributes.memory").String()),
		SystemDiskSize:    20,
		EphemeralDiskSize: decodeStorage(gjson.GetBytes(js, "product.attributes.storage").String()),
		CPUSpeed:          decodeClockSpeed(gjson.GetBytes(js, "product.attributes.clockSpeed").String()),
		CreatedAt:         creationDate,
		OneDemandPrice:    readPrice(js),
	}
	return tpl
}

//List returns available VM templates
func (mgr *ServerTemplateManager) List() ([]api.ServerTemplate, error) {
	out, err := mgr.AWS.PricingClient.GetProducts(&pricing.GetProductsInput{
		Filters: []*pricing.Filter{
			{
				Field: aws.String("ServiceCode"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("AmazonEC2"),
			},
			{
				Field: aws.String("location"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("US East (Ohio)"),
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
		},
		FormatVersion: aws.String("aws_v1"),
		ServiceCode:   aws.String("AmazonEC2"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error listing server templates")
	}
	var result []api.ServerTemplate
	for _, price := range out.PriceList {
		tpl := toTemplate(price)
		if tpl.RAMSize == 0 {
			continue
		}
		result = append(result, *tpl)

	}
	return result, nil
}

//Get returns the template identified by ids
func (mgr *ServerTemplateManager) Get(id string) (*api.ServerTemplate, error) {
	out, err := mgr.AWS.PricingClient.GetProducts(&pricing.GetProductsInput{
		Filters: []*pricing.Filter{
			{
				Field: aws.String("ServiceCode"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("AmazonEC2"),
			},
			{
				Field: aws.String("location"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("US East (Ohio)"),
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
				Field: aws.String("sku"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String(id),
			},
		},
		FormatVersion: aws.String("aws_v1"),
		ServiceCode:   aws.String("AmazonEC2"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error listing server templates")
	}

	for _, price := range out.PriceList {
		res := toTemplate(price)
		if res.ID == id {
			return res, nil
		}

	}
	return nil, fmt.Errorf("error getting Server Template: Server template %s not found", id)
}
