package hetznerdns

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/germanbrew/terraform-provider-hetznerdns/hetznerdns/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRecord() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRecordCreate,
		ReadContext:   resourceRecordRead,
		UpdateContext: resourceRecordUpdate,
		DeleteContext: resourceRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceRecordCreate(c context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Updating resource record")
	client := m.(*api.Client)

	zoneID, zoneIDNonEmpty := d.GetOk("zone_id")
	if !zoneIDNonEmpty {
		return diag.Errorf("Zone ID of record not set")
	}

	name, nameNonEmpty := d.GetOk("name")
	if !nameNonEmpty {
		return diag.Errorf("Name of record not set")
	}

	recordType, typeNonEmpty := d.GetOk("type")
	if !typeNonEmpty {
		return diag.Errorf("Type of record not set")
	}

	value, valueNonEmpty := d.GetOk("value")
	if !valueNonEmpty {
		return diag.Errorf("Value of record not set")
	}

	if recordType.(string) == "TXT" {
		value = prepareTXTRecordValue(value.(string))
	}

	opts := api.CreateRecordOpts{
		ZoneID: zoneID.(string),
		Name:   name.(string),
		Type:   recordType.(string),
		Value:  value.(string),
	}

	tTL, tTLNonEmpty := d.GetOk("ttl")
	if tTLNonEmpty {
		nonEmptyTTL := tTL.(int)
		opts.TTL = &nonEmptyTTL
	}

	record, err := client.CreateRecord(opts)
	if err != nil {
		log.Printf("[ERROR] Error creating DNS record %s: %s", opts.Name, err)
		return diag.Errorf("Error creating DNS record %s: %s", opts.Name, err)
	}

	d.SetId(record.ID)
	return resourceRecordRead(c, d, m)
}

func resourceRecordRead(c context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Reading resource record")
	client := m.(*api.Client)

	id := d.Id()
	record, err := client.GetRecord(id)
	if err != nil {
		return diag.Errorf("Error getting record with id %s: %s", id, err)
	}

	if record == nil {
		log.Printf("[WARN] DNS record with id %s doesn't exist, removing it from state", id)
		d.SetId("")
		return nil
	}

	if record.Type == "TXT" {
		if strings.HasPrefix(record.Value, "\"") && strings.HasSuffix(record.Value, "\" ") {
			record.Value = unescapeTXTRecordValue(record.Value)
		}
	}

	d.SetId(record.ID)
	d.Set("name", record.Name)
	d.Set("zone_id", record.ZoneID)
	d.Set("type", record.Type)

	d.Set("ttl", nil)
	if record.HasTTL() {
		d.Set("ttl", record.TTL)
	}
	d.Set("value", record.Value)

	return nil
}

func resourceRecordUpdate(c context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Updating resource record")
	client := m.(*api.Client)

	id := d.Id()
	record, err := client.GetRecord(id)
	if err != nil {
		return diag.Errorf("Error getting record with id %s: %s", id, err)
	}

	if record == nil {
		log.Printf("[WARN] DNS record with id %s doesn't exist, removing it from state", id)
		d.SetId("")
		return nil
	}

	if record.Type == "TXT" {
		// Unescape the TXT record value if it is escaped
		if strings.HasPrefix(record.Value, "\"") && strings.HasSuffix(record.Value, "\" ") {
			record.Value = unescapeTXTRecordValue(record.Value)
		}
	}

	if d.HasChanges("name", "ttl", "type", "value") {
		record.Name = d.Get("name").(string)

		record.TTL = nil
		ttl, ttlNonEmpty := d.GetOk("ttl")
		if ttlNonEmpty {
			ttl := ttl.(int)
			record.TTL = &ttl
		}
		record.Type = d.Get("type").(string)
		record.Value = d.Get("value").(string)
		_, err = client.UpdateRecord(*record)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceRecordRead(c, d, m)
}

func resourceRecordDelete(c context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Deleting resource record")

	client := m.(*api.Client)
	recordID := d.Id()

	err := client.DeleteRecord(recordID)
	if err != nil {
		log.Printf("[ERROR] Error deleting record %s: %s", recordID, err)
		return diag.FromErr(err)
	}

	return nil
}

// If the value in a TXT record is longer than 255 bytes, it needs to be split into multiple parts.
// Each part needs to be enclosed in double quotes and separated by a space.
// https://datatracker.ietf.org/doc/html/rfc4408#section-3.1.3
func prepareTXTRecordValue(value string) string {
	if len(value) < 255 {
		log.Printf("[DEBUG] TXT record value is shorter than 255 bytes, no need to split it")
		return value
	}

	// If the String is already in the correct format, return it as is
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		log.Printf("[DEBUG] TXT record value is already in the correct format")
		return value
	}

	// Split the DKIM key into 255 byte parts
	parts := splitStringBy255Bytes(value)
	for i, part := range parts {
		parts[i] = "\"" + part + "\""
	}

	log.Printf("[DEBUG] TXT record value has been split into %d parts", len(parts))
	log.Print(parts)
	return strings.Join(parts, " ")
}

// Strings in TXT records are by design limited to 255 bytes.
// Strings with more characters get split to substrings separated by space every 255 bytes/characters.
// https://datatracker.ietf.org/doc/html/rfc4408#section-3.1.3
func splitStringBy255Bytes(value string) []string {
	var parts []string
	for len(value) > 0 {
		if len(value) > 255 {
			parts = append(parts, value[:255])
			value = value[255:]
		} else {
			parts = append(parts, value)
			break
		}
	}
	return parts
}

func unescapeTXTRecordValue(value string) string {
	value = strings.ReplaceAll(value, "\" ", "")
	value = strings.ReplaceAll(value, "\"", "")
	return strings.TrimSpace(value)
}
