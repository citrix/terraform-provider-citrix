// Copyright © 2024. Citrix Systems, Inc.

package validators

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.String = IPFilterValidator{}
)

// IpFilterValidator validates if.
type IPFilterValidator struct{}

// Description implements validator.String.
func (v IPFilterValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

// MarkdownDescription implements validator.String.
func (v IPFilterValidator) MarkdownDescription(context.Context) string {
	return "value must be a valid IP address, IP address range, or IP address CIDR block"
}

func (v IPFilterValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	ipAddress := req.ConfigValue.ValueString()
	err := ipAddressValidation(ipAddress)
	if err != nil {
		errMsg := err.Error()
		if err.Error() == "" {
			errMsg = v.Description(ctx)
		}
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			errMsg,
			ipAddress,
		))
	}
}

func ipAddressValidation(ipAddress string) error {
	ipAddresses := strings.Split(ipAddress, "-")
	if len(ipAddresses) == 2 {
		// Working with an IP address range
		err := ipAddressValidationHelper(ipAddresses[0])
		if err != nil {
			return err
		}

		err = ipAddressValidationHelper(ipAddresses[1])
		if err != nil {
			return err
		}

		// Make sure the first ip address is before the second ip address.
		firstIpAddress, err := convertIpAddressToNumber(ipAddresses[0])
		if err != nil {
			return err
		}
		secondIpAddress, err := convertIpAddressToNumber(ipAddresses[1])
		if err != nil {
			return err
		}
		if firstIpAddress >= secondIpAddress {
			return fmt.Errorf("the first IP address `%s` must be before the second IP address `%s`", ipAddresses[0], ipAddresses[1])
		}
		return nil
	}
	// working with a single IP Address
	return ipAddressValidationHelper(ipAddress)
}

func ipAddressValidationHelper(ipInput string) error {
	if _, ipNet, err := net.ParseCIDR(ipInput); err == nil {
		mask, _ := ipNet.Mask.Size()
		if mask < 0 || mask > 39 {
			return fmt.Errorf("the IP address masks lower than 0 or greater than 39 are not supported")
		}
		return nil
	} else if ipAddressWithoutCidr := net.ParseIP(ipInput); ipAddressWithoutCidr != nil {
		return nil
	} else {
		isIpv4 := !strings.Contains(ipInput, ":")
		hasAsterisk := strings.Contains(ipInput, "*")
		if hasAsterisk && len(strings.Split(ipInput, "/")) == 2 {
			return fmt.Errorf("the IP address range with asterisk wildcard cannot be used with CIDR notation")
		}
		if hasAsterisk && isIpv4 {
			if octets := strings.Split(ipInput, "."); len(octets) == 4 {
				// Must process each part individually; the asterisk wildcard can only occur after a numbered element, not between.
				// 12.*.12.4                         Not Valid
				// 12.*.*.*                          Valid
				// 12.45.*.*                         Valid
				// 12.23.76.*                        Valid
				wildcardFound := false
				for _, octet := range octets {
					if octet == "*" {
						wildcardFound = true
					} else if !wildcardFound {
						octetInt, err := strconv.Atoi(octet)
						if err != nil {
							return fmt.Errorf("the IP address octet `%s` is not a valid integer", octet)
						}
						if octetInt < 0 || octetInt > 255 {
							return fmt.Errorf("invalid IP address octet `%s`. Each IP address octet cannot be lower than 0 or greater than 255", octet)
						}
					} else {
						return fmt.Errorf("the asterisk wildcard can only occur after a numbered element, not between")
					}
				}
				return nil
			}
		}
	}

	return fmt.Errorf("value must be a valid IP address, IP address range, or IP address CIDR block")
}

func convertIpAddressToNumber(ipAddress string) (int, error) {
	byteIP := strings.Split(ipAddress, ".")
	byteIp1, err := strconv.Atoi(byteIP[0])
	if err != nil {
		return -1, err
	}
	byteIp2, err := strconv.Atoi(byteIP[1])
	if err != nil {
		return -1, err
	}
	byteIp3, err := strconv.Atoi(byteIP[2])
	if err != nil {
		return -1, err
	}
	byteIp4, err := strconv.Atoi(byteIP[3])
	if err != nil {
		return -1, err
	}

	ip := byteIp1 * 16777216 // 256 * 256 * 256
	ip += byteIp2 * 65536    // 256 * 256
	ip += byteIp3 * 256
	ip += byteIp4

	return ip, nil
}

// ValidateIpFilter checks that a set of path.Expression has a non-null value,
// if the current attribute or block is set to one of the values defined in onValues array.
//
// Relative path.Expression will be resolved using the attribute or block
// being validated.
func ValidateIPFilter() IPFilterValidator {
	return IPFilterValidator{}
}
