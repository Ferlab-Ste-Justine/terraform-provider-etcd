package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourcePrefixRangeEnd() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrefixRangeEndRead,
		Schema: map[string]*schema.Schema{
			"key": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"range_end": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePrefixRangeEndRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	key := d.Get("key").(string)
	
	d.SetId(key)
	
	rangeEnd, err := getPrefixRangeEnd(key)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}

	d.Set("range_end", rangeEnd)
	
	return nil
}