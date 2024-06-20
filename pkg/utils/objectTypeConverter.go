package utils

func ObjectTypeConverter(objectType string) string {
	switch objectType {
	case "tp_objects":
		return "ТП"
	case "kl04kv_objects":
		return "КЛ 04 КВ"
	case "mjd_objects":
		return "МЖД"
	case "sip_objects":
		return "СИП"
	case "stvt_objects":
		return "СТВТ"
	default:
		return "Другое"
	}
}
