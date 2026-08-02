package pmocker

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

func readAttrString(entityID uint, key string) string {
	var attr pmocker.PMAttr
	global.GVA_DB.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValString != nil {
		return *attr.ValString
	}
	if attr.ValDate != nil {
		return *attr.ValDate
	}
	if attr.ValDateTime != nil {
		return *attr.ValDateTime
	}
	if attr.ValJSON != nil {
		return *attr.ValJSON
	}
	return ""
}

func readAttrDecimal(entityID uint, key string) float64 {
	var attr pmocker.PMAttr
	global.GVA_DB.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValDecimal != nil {
		return *attr.ValDecimal
	}
	return 0
}

func readAttrInt(entityID uint, key string) int64 {
	var attr pmocker.PMAttr
	global.GVA_DB.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValInt != nil {
		return *attr.ValInt
	}
	return 0
}

func readAttrRef(entityID uint, key string) uint {
	var attr pmocker.PMAttr
	global.GVA_DB.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValRef != nil {
		return *attr.ValRef
	}
	if attr.ValInt != nil {
		return uint(*attr.ValInt)
	}
	return 0
}

func writeAttrString(entityID uint, key, value string) error {
	v := value
	return upsertAttr(entityID, key, &v, nil, nil)
}

func writeAttrDecimal(entityID uint, key string, value float64) error {
	v := value
	return upsertAttr(entityID, key, nil, &v, nil)
}

func upsertAttr(entityID uint, key string, str *string, dec *float64, intv *int64) error {
	var attr pmocker.PMAttr
	err := global.GVA_DB.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr).Error
	if err != nil {
		attr = pmocker.PMAttr{EntityID: entityID, FieldKey: key, ValString: str, ValDecimal: dec, ValInt: intv}
		return global.GVA_DB.Create(&attr).Error
	}
	updates := map[string]interface{}{}
	if str != nil {
		updates["val_string"] = *str
	}
	if dec != nil {
		updates["val_decimal"] = *dec
	}
	if intv != nil {
		updates["val_int"] = *intv
	}
	if len(updates) == 0 {
		return nil
	}
	return global.GVA_DB.Model(&pmocker.PMAttr{}).Where("id = ?", attr.ID).Updates(updates).Error
}

func attrValueString(attr pmocker.PMAttr) string {
	if attr.ValString != nil {
		return *attr.ValString
	}
	if attr.ValDecimal != nil {
		return strconv.FormatFloat(*attr.ValDecimal, 'f', -1, 64)
	}
	if attr.ValInt != nil {
		return strconv.FormatInt(*attr.ValInt, 10)
	}
	if attr.ValDate != nil {
		return *attr.ValDate
	}
	if attr.ValDateTime != nil {
		return *attr.ValDateTime
	}
	if attr.ValBool != nil {
		if *attr.ValBool {
			return "true"
		}
		return "false"
	}
	if attr.ValJSON != nil {
		return *attr.ValJSON
	}
	if attr.ValRef != nil {
		return strconv.FormatUint(uint64(*attr.ValRef), 10)
	}
	return ""
}
