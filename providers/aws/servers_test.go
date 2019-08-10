package aws_test

import (
	"fmt"
	"io/ioutil"
	"sort"
	"testing"
	"time"

	"github.com/tidwall/gjson"
)

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

func TestReadPrice(t *testing.T) {
	price, _ := ioutil.ReadFile("/tmp/m5d.24xlarge.json")
	m := gjson.GetBytes(price, "terms.OnDemand").Map()

	plist := PriceList{}
	for _, res := range m {
		mp := res.Map()
		t, _ := time.Parse(time.RFC3339, mp["effectiveDate"].String())
		dim := mp["priceDimensions"]
		value := float32(0.0)
		for _, res2 := range dim.Map() {
			value = float32(res2.Get("pricePerUnit.USD").Float())
		}
		p := Price{
			date:  t,
			value: value,
		}
		plist = append(plist, p)
	}
	sort.Sort(plist)
	p := plist[0]
	fmt.Printf("%+v\n", p)
}
