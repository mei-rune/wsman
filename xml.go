package wsman

import (
	"encoding/xml"
	"errors"
	"fmt"
)

func ElementNotExists(nm string) error {
	return errors.New("'" + nm + "' is not exists.")
}

type Reference struct {
	// <a:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
	// <a:ReferenceParameters>
	//     <w:ResourceURI>http://schemas.microsoft.com/wbem/wsman/1/wmi/root/virtualization/Msvm_Synthetic3DDisplayControllerSettingData</w:ResourceURI>
	//     <w:SelectorSet>
	//         <w:Selector Name="InstanceID">Microsoft:Definition\06FF76FA-2D58-4BAF-9F8D-455773824F37\Default</w:Selector>
	//     </w:SelectorSet>
	// </a:ReferenceParameters>

	Address     string
	ResourceURI string
	SelectorSet map[string]string
}

func readXmlText(decoder *xml.Decoder) (string, error) {
	var context string
	for {
		token, err := decoder.Token()
		if nil != err {
			return context, err
		}
		switch v := token.(type) {
		case xml.EndElement:
			return context, nil
		case xml.CharData:
			context = string(v)
		// case xml.StartElement:
		//  switch v.Name.Local {
		//  case "Datetime":
		//    txt, e := readXmlText(decoder)
		//    if nil != e {
		//      return "", e
		//    }
		//    if e = exitElement(decoder, 0); nil != e {
		//      return txt, e
		//    }
		//    return txt, nil
		//  default:
		//    return context, errors.New("element '" + v.Name.Local + "' is not excepted, excepted is CharData")
		//  }
		default:
			return context, fmt.Errorf("token '%T' is not excepted, excepted is CharData", v)
		}
	}
}

func readReferenceParameters(decoder *xml.Decoder, reference *Reference) error {
	// <a:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
	// <a:ReferenceParameters>
	//     <w:ResourceURI>http://schemas.microsoft.com/wbem/wsman/1/wmi/root/virtualization/Msvm_Synthetic3DDisplayControllerSettingData</w:ResourceURI>
	//     <w:SelectorSet>
	//         <w:Selector Name="InstanceID">Microsoft:Definition\06FF76FA-2D58-4BAF-9F8D-455773824F37\Default</w:Selector>
	//     </w:SelectorSet>
	// </a:ReferenceParameters>

	inSet := false
	for {
		token, err := decoder.Token()
		if nil != err {
			return err
		}
		switch v := token.(type) {
		case xml.EndElement:
			if inSet {
				inSet = false
				break
			}
			return nil
		case xml.CharData:
		case xml.StartElement:
			switch v.Name.Local {
			case "ResourceURI":
				txt, e := readXmlText(decoder)
				if nil != e {
					return e
				}
				reference.ResourceURI = txt
			case "SelectorSet":
				inSet = true
			case "Selector":
				if !inSet {
					return errors.New("element '" + v.Name.Local + "' is not excepted, excepted is ResourceURI or SelectorSet")
				}

				txt, e := readXmlText(decoder)
				if nil != e {
					return e
				}

				for _, attr := range v.Attr {
					if "Name" == attr.Name.Local {
						if reference.SelectorSet == nil {
							reference.SelectorSet = map[string]string{}
						}
						reference.SelectorSet[attr.Value] = txt
					}
				}
			default:
				return errors.New("element '" + v.Name.Local + "' is not excepted, excepted is CharData")
			}
		default:
			return fmt.Errorf("token '%T' is not excepted, excepted is CharData", v)
		}
	}
}

func readXmlValue(decoder *xml.Decoder) (interface{}, error) {
	var reference *Reference
	var charData string
	for {
		token, err := decoder.Token()
		if nil != err {
			return nil, err
		}
		switch v := token.(type) {
		case xml.EndElement:
			if reference != nil {
				return reference, nil
			}
			if charData != "" {
				return charData, nil
			}
			return nil, nil
		case xml.CharData:
			charData = string(v)
		case xml.StartElement:
			switch v.Name.Local {
			case "Datetime":
				txt, e := readXmlText(decoder)
				if nil != e {
					return "", e
				}
				if e = exitElement(decoder, 0); e != nil {
					return nil, e
				}
				return txt, nil
			case "Interval":
				txt, e := readXmlText(decoder)
				if nil != e {
					return "", e
				}
				if e = exitElement(decoder, 0); e != nil {
					return nil, e
				}
				return txt, nil
			case "Address":
				address, e := readXmlText(decoder)
				if nil != e {
					return nil, e
				}
				if reference == nil {
					reference = &Reference{}
				}
				reference.Address = address
			case "ReferenceParameters":
				if reference == nil {
					reference = &Reference{}
				}
				if e := readReferenceParameters(decoder, reference); e != nil {
					return nil, e
				}
			default:
				return nil, errors.New("element '" + v.Name.Local + "' is not excepted, excepted is CharData")
			}
		default:
			return nil, fmt.Errorf("token '%T' is not excepted, excepted is CharData", v)
		}
	}
}

func locateElement(decoder *xml.Decoder, nm string) (bool, error) {
	depth := 0
	for {
		t, err := decoder.Token()
		if nil != err {
			return false, err
		}
		switch t := t.(type) {
		case xml.EndElement:
			depth--
			if depth < 0 {
				return false, nil
			}
		case xml.StartElement:
			if 0 == depth && nm == t.Name.Local {
				return true, nil
			}
			depth++
		}
	}
}

func nextElement(decoder *xml.Decoder) (xml.Name, []xml.Attr, error) {
	for {
		t, err := decoder.Token()
		if nil != err {
			return EMPTY_NAME, nil, err
		}
		switch el := t.(type) {
		case xml.EndElement:
			return el.Name, nil, ElementEndError
		case xml.StartElement:
			return el.Name, el.Attr, nil
		}
	}
}

func exitElement(decoder *xml.Decoder, depth int) error {
	for {
		t, err := decoder.Token()
		if nil != err {
			return err
		}
		switch v := t.(type) {
		case xml.EndElement:
			depth--
			if depth < 0 {
				return nil
			}
		case xml.StartElement:
			return errors.New("StartElement '" + v.Name.Local + "' is not excepted, excepted is EndElement")
		}
	}
}

func skipElement(decoder *xml.Decoder, depth int) error {
	for {
		t, err := decoder.Token()
		if nil != err {
			return err
		}
		switch t.(type) {
		case xml.EndElement:
			depth--
			if depth < 0 {
				return nil
			}
		case xml.StartElement:
			depth++
		}
	}
}

func locateElements(decoder *xml.Decoder, names []string) (bool, error) {
	for _, nm := range names {
		ok, e := locateElement(decoder, nm)
		if nil != e {
			return false, e
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func isNil(attrs []xml.Attr) bool {
	for _, attr := range attrs {
		if "nil" == attr.Name.Local && "true" == attr.Value {
			return true
		}
	}
	return false
}

func toMap(decoder *xml.Decoder) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	for {
		t, err := decoder.Token()
		if nil != err {
			return nil, err
		}
		switch v := t.(type) {
		case xml.EndElement:
			return m, nil
		case xml.StartElement:
			var value interface{}
			if isNil(v.Attr) {
				if e := skipElement(decoder, 0); nil != e {
					return nil, e
				}
				value = nil
			} else {
				xmlValue, e := readXmlValue(decoder)
				if nil != e {
					if ElementEndError != e {
						return nil, e
					}
					value = nil
				} else {
					if xmlValue == nil && len(v.Attr) > 0 {
						for _, attr := range v.Attr {
							if "SystemTime" == attr.Name.Local {
								xmlValue = attr.Value
							}
						}
					}
					value = xmlValue
				}
			}

			old, ok := m[v.Name.Local]
			if !ok {
				m[v.Name.Local] = value
			} else if aa, ok := old.([]interface{}); ok {
				m[v.Name.Local] = append(aa, value)
			} else {
				m[v.Name.Local] = []interface{}{old, value}
			}
		}
	}
}
