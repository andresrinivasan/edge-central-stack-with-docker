// Copyright (C) 2020-2021 IOTech Ltd

package xpert

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/cache"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

type Filter struct{}

// NewFilter creates, initializes and returns a new instance of Filter
func NewFilter() Filter {
	return Filter{}
}

// FilterByValueMaxMin removes readings with outlier value that is out of maximum and minimum as defined in DeviceResource.
// This function only targets on numeric value, e.g. Int, Float, for non-numeric readings, the Filter simply ignores.
// This function also only filter when Maximum/Minimum is defined in the DeviceResource; when only Maximum is defined,
// compare with Maximum only; when only Minimum is defined, compare with Minimum only; when both Maximum and Minimum are
// defined, only the readings with numeric values between Maximum and Minimum will be allowed to pass to next function
// in the pipeline.  Please note that both FilterValues and FilterOut parameters doesn't take effect to this function.
// This function will return an error and stop the pipeline if a non-edgex event is received or if no data is received.
func (f Filter) FilterByValueMaxMin(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, result interface{}) {

	ctx.LoggingClient().Debug("Filtering by DeviceResource Maximum and Minimum")

	if data == nil {
		return false, errors.New("no Event Received")
	}

	// FilterByValueMaxMin expects to deal with event model; for non-event model, simply throw out error.
	existingEvent, ok := data.(dtos.Event)
	if !ok {
		return false, errors.New("type received is not an Event")
	}

	// auxEvent will be the event passed into pipeline
	auxEvent := dtos.Event{
		DeviceName: existingEvent.DeviceName,
		Origin:     existingEvent.Origin,
		Readings:   []dtos.BaseReading{},
	}

	// iterate each reading to examine if its value is an outlier (bigger than max or smaller than min)
	for _, reading := range existingEvent.Readings {
		// get the DeviceResource from cache
		dr, ok := cache.DeviceResources(ctx).ForName(reading.ProfileName, reading.ResourceName)
		if !ok {
			ctx.LoggingClient().Warnf("No corresponding DeviceResource for reading %s.  Ignore by FilterByValueMaxMin", reading.Id)
		} else {
			result, err := f.isOutlier(reading, dr)
			if err != nil {
				ctx.LoggingClient().Errorf("encountering error while if reading %s is outlier.  Ignore by FilterByValueMaxMin", reading.Id)
				continue
			}
			if result {
				ctx.LoggingClient().Debugf("Drops outlier reading value: %s, drop", reading.Id)
				continue
			}
			ctx.LoggingClient().Tracef("Reading accepted: %s", reading.Id)
			// accepted reading (value is not outlier) will be appended to the readings of auxEvent
			auxEvent.Readings = append(auxEvent.Readings, reading)
		}
	}
	thereExistReadings := len(auxEvent.Readings) > 0
	var returnResult dtos.Event
	if thereExistReadings {
		returnResult = auxEvent
	}
	return thereExistReadings, returnResult
}

func (f Filter) isOutlier(reading dtos.BaseReading, dr dtos.DeviceResource) (bool, error) {
	// Convenience short cut
	drp := dr.Properties

	//only deal with reading whose ValueType is identical to the passed-in DeviceResource's Type
	if reading.ValueType == drp.ValueType {
		//only deal with numeric value
		switch drp.ValueType {
		case common.ValueTypeUint8:
			return f.isOutlierInteger(reading.Value, drp.Maximum, drp.Minimum, 8, false)
		case common.ValueTypeUint16:
			return f.isOutlierInteger(reading.Value, drp.Maximum, drp.Minimum, 16, false)
		case common.ValueTypeUint32:
			return f.isOutlierInteger(reading.Value, drp.Maximum, drp.Minimum, 32, false)
		case common.ValueTypeUint64:
			return f.isOutlierInteger(reading.Value, drp.Maximum, drp.Minimum, 64, false)
		case common.ValueTypeInt8:
			return f.isOutlierInteger(reading.Value, drp.Maximum, drp.Minimum, 8, true)
		case common.ValueTypeInt16:
			return f.isOutlierInteger(reading.Value, drp.Maximum, drp.Minimum, 16, true)
		case common.ValueTypeInt32:
			return f.isOutlierInteger(reading.Value, drp.Maximum, drp.Minimum, 32, true)
		case common.ValueTypeInt64:
			return f.isOutlierInteger(reading.Value, drp.Maximum, drp.Minimum, 64, true)
		case common.ValueTypeFloat32:
			return f.isFloatOutlier(reading.Value, drp.Maximum, drp.Minimum, 32)
		case common.ValueTypeFloat64:
			return f.isFloatOutlier(reading.Value, drp.Maximum, drp.Minimum, 64)
		default:
			return false, nil
		}
	}
	return false, nil
}

// isOutlierInteger determines if readingValueStr could be parsed into integer and if its value is an outlier (bigger than max or smaller than min)
func (f Filter) isOutlierInteger(readingValueStr string, max interface{}, min interface{}, bitSize int, signed bool) (bool, error) {
	if len(readingValueStr) == 0 {
		return true, fmt.Errorf("zero-length reading value found")
	}
	if signed {
		// parse readingValueStr to integer first, if errors, return false and error message as we couldn't determine if it's an outlier
		readingValue, err := strconv.ParseInt(readingValueStr, 10, bitSize)
		if err != nil {
			return false, fmt.Errorf("failed to parse non-empty reading value %s to integer with bitSize %v. Error:%s", readingValueStr, bitSize, err)
		}
		// compare with max if max is specified in device resource
		if maxStr, ok := max.(string); ok && len(maxStr) > 0 {
			maxValue, err := strconv.ParseInt(maxStr, 10, bitSize)
			if err != nil {
				return false, fmt.Errorf("failed to parse non-empty maximum value %s to integer with bitSize %v. Error:%s", maxStr, bitSize, err)
			}
			if readingValue > maxValue {
				return true, nil
			}
		}
		// compare with min if min is specified in device resource
		if minStr, ok := min.(string); ok && len(minStr) > 0 {
			minValue, err := strconv.ParseInt(minStr, 10, bitSize)
			if err != nil {
				return false, fmt.Errorf("failed to parse non-empty minimum value %s to integer with bitSize %v. Error:%s", minStr, bitSize, err)
			}
			if readingValue < minValue {
				return true, nil
			}
		}
	} else {
		// parse readingValueStr to unsigned integer first, if errors, return false and error message as we couldn't determine if it's an outlier
		readingValue, err := strconv.ParseUint(readingValueStr, 10, bitSize)
		if err != nil {
			return false, fmt.Errorf("failed to parse non-empty reading value %s to unsigned integer with bitSize %v. Error:%s", readingValueStr, bitSize, err)
		}
		// compare with max if max is specified in device resource
		if maxStr, ok := max.(string); ok && len(maxStr) > 0 {
			maxValue, err := strconv.ParseUint(maxStr, 10, bitSize)
			if err != nil {
				return false, fmt.Errorf("failed to parse non-empty maximum value %s to unsigned integer with bitSize %v. Error:%s", maxStr, bitSize, err)
			}
			if readingValue > maxValue {
				return true, nil
			}
		}
		// compare with min if min is specified in device resource
		if minStr, ok := min.(string); ok && len(minStr) > 0 {
			minValue, err := strconv.ParseUint(minStr, 10, bitSize)
			if err != nil {
				return false, fmt.Errorf("failed to parse non-empty minimum value %s to unsigned integer with bitSize %v. Error:%s", minStr, bitSize, err)
			}
			if readingValue < minValue {
				return true, nil
			}
		}
	}
	return false, nil
}

// isFloatOutlier determines if readingValueStr could be parsed into float32 or float64 and if its value is a outlier (bigger than max or smaller than min)
func (f Filter) isFloatOutlier(readingValueStr string, max interface{}, min interface{}, bitSize int) (bool, error) {
	floatValue, err := strconv.ParseFloat(readingValueStr, bitSize)
	if err != nil {
		return false, fmt.Errorf("failed to parse reading value '%s' to float with bitSize %v. Error:%s", readingValueStr, bitSize, err)
	}
	// compare with max if max is specified in device resource
	if maxStr, ok := max.(string); ok && len(maxStr) > 0 {
		maxValue, err := strconv.ParseFloat(maxStr, bitSize)
		if err != nil {
			return false, fmt.Errorf("failed to parse non-empty maximum value '%s' to float with bitSize %v. Error:%s", maxStr, bitSize, err)
		}
		if floatValue > maxValue {
			return true, nil
		}
	}
	// compare with min if min is specified in device resource
	if minStr, ok := min.(string); ok && len(minStr) > 0 {
		minValue, err := strconv.ParseFloat(minStr, bitSize)
		if err != nil {
			return false, fmt.Errorf("failed to parse non-empty minimum value '%s' to float with bitSize %v. Error:%s", minStr, bitSize, err)
		}
		if floatValue < minValue {
			return true, nil
		}
	}
	return false, nil
}
