package rule_algorithm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"../confreader"
	"../my_db"
	"github.com/astaxie/beego"
)

const TIME_LAYOUT = "2006-01-02 15:04:05"

func GetAttendanceResult(rule my_db.RuleInfo, strTime string) (string, error) {
	if rule.Late_time < 0 {
		rule.Late_time = 0
	}

	if rule.Left_early < 0 {
		rule.Left_early = 0
	}

	timeSplitArray := strings.Split(strTime, "T")
	if len(timeSplitArray) != 2 {
		return "", errors.New("错误的创建时间")
	}

	rule.Work_begin = timeSplitArray[0] + " " + rule.Work_begin
	rule.Work_end = timeSplitArray[0] + " " + rule.Work_end
	rule.Offwork_begin = timeSplitArray[0] + " " + rule.Offwork_begin
	rule.Offwork_end = timeSplitArray[0] + " " + rule.Offwork_end

	//fmt.Println("时间数据:", rule.Work_begin, rule.Work_end, rule.Offwork_begin, rule.Offwork_end)

	timeWorkBegin, err := time.Parse(TIME_LAYOUT, rule.Work_begin)
	if err != nil {
		return "", err
	}

	timeWorkEnd, err := time.Parse(TIME_LAYOUT, rule.Work_end)
	if err != nil {
		return "", err
	}

	timeOffworkBegin, err := time.Parse(TIME_LAYOUT, rule.Offwork_begin)
	if err != nil {
		return "", err
	}

	timeOffworkEnd, err := time.Parse(TIME_LAYOUT, rule.Offwork_end)
	if err != nil {
		return "", err
	}

	timeSplitArray_ := strings.Split(timeSplitArray[1], "Z")

	strTime = timeSplitArray[0] + " " + timeSplitArray_[0]
	timeAttendance, err := time.Parse(TIME_LAYOUT, strTime)
	if err != nil {
		return "", err
	}

	//在上班打卡时间段   属于正常打卡
	lateSpan, err := time.ParseDuration(fmt.Sprintf("%dm", rule.Late_time))
	if timeAttendance.After(timeWorkBegin) && timeAttendance.Before(timeWorkEnd) {
		return confreader.GetValue("Work"), nil
	}

	earlySpan, err := time.ParseDuration(fmt.Sprintf("-%dm", rule.Left_early))
	if timeAttendance.After(timeOffworkBegin) && timeAttendance.Before(timeOffworkEnd) {
		return confreader.GetValue("Offwork"), nil
	}

	if rule.Late_time > 0 && rule.Left_early > 0 {
		//严重迟到或早退
		if timeAttendance.After(timeWorkEnd.Add(lateSpan)) && timeAttendance.Before(timeOffworkBegin.Add(earlySpan)) {
			strTimeNoon := timeSplitArray[0] + " " + "12:00:00"
			timeNoon, err := time.Parse(TIME_LAYOUT, strTimeNoon)
			if err != nil {
				return "", err
			}
			if timeAttendance.Before(timeNoon) {
				return confreader.GetValue("SeriouslyLate"), nil
			} else {
				return confreader.GetValue("SevereEarlyLeave"), nil
			}
		}
	}

	if timeAttendance.After(timeWorkEnd) && timeAttendance.Before(timeOffworkBegin) {
		strTimeNoon := timeSplitArray[0] + " " + "12:00:00"
		timeNoon, err := time.Parse(TIME_LAYOUT, strTimeNoon)
		if err != nil {
			return "", err
		}
		if timeAttendance.Before(timeNoon) {
			return confreader.GetValue("Late"), nil
		} else {
			return confreader.GetValue("EarlyLeave"), nil
		}

	}

	return confreader.GetValue("NonAttendanceTimePeriod"), nil
}

//GetAttendanceResultByRangeInfo __
func GetAttendanceResultByRangeInfo(rule my_db.RangeInfo, strTime string) (string, string, error) {
	if rule.LateTime < 0 {
		rule.LateTime = 0
	}

	if rule.LeftEarly < 0 {
		rule.LeftEarly = 0
	}

	arrayBegin := strings.Split(rule.Begin, " ")
	rule.Begin = arrayBegin[0]

	strWorkTime := rule.WorkBegin + "-" + rule.WorkEnd
	strOffWorkTime := rule.OffWorkBegin + "-" + rule.OffWorkEnd

	rule.Time1 = rule.Begin + " " + rule.Time1
	rule.Time2 = rule.Begin + " " + rule.Time2
	rule.WorkBegin = rule.Begin + " " + rule.WorkBegin
	rule.WorkEnd = rule.Begin + " " + rule.WorkEnd
	rule.OffWorkBegin = rule.Begin + " " + rule.OffWorkBegin
	rule.OffWorkEnd = rule.Begin + " " + rule.OffWorkEnd

	//fmt.Println("时间数据:", rule.Work_begin, rule.Work_end, rule.Offwork_begin, rule.Offwork_end)

	timeWorkBegin, err := time.Parse(TIME_LAYOUT, rule.WorkBegin+":00")
	if err != nil {
		return "", "", err
	}

	timeWorkEnd, err := time.Parse(TIME_LAYOUT, rule.WorkEnd+":00")
	if err != nil {
		return "", "", err
	}

	timeOffworkBegin, err := time.Parse(TIME_LAYOUT, rule.OffWorkBegin+":00")
	if err != nil {
		return "", "", err
	}

	timeOffworkEnd, err := time.Parse(TIME_LAYOUT, rule.OffWorkEnd+":00")
	if err != nil {
		return "", "", err
	}

	timeTime1, err := time.Parse(TIME_LAYOUT, rule.Time1+":00")
	if err != nil {
		return "", "", err
	}

	timeTime2, err := time.Parse(TIME_LAYOUT, rule.Time2+":00")
	if err != nil {
		return "", "", err
	}

	timeSpan := timeTime2.Sub(timeTime1)
	nHour := timeSpan.Hours()
	nHour = nHour / 2
	timeSpan, err = time.ParseDuration(fmt.Sprintf("%f", nHour) + "h")
	if err != nil {
		return "", "", err
	}

	timeMidle := timeTime1.Add(timeSpan)

	//fmt.Println("中间时间段:", timeMidle.Format(TIME_LAYOUT))

	timeAttendance, err := time.Parse(TIME_LAYOUT, strTime)
	if err != nil {
		return "", "", err
	}

	//在上班打卡时间段   属于正常打卡
	lateSpan, err := time.ParseDuration(fmt.Sprintf("%dm", rule.LateTime))
	if timeAttendance.After(timeWorkBegin) && timeAttendance.Before(timeWorkEnd) && rule.WorkCheck > 0 {
		return confreader.GetValue("Work"), strWorkTime, nil
	}

	earlySpan, err := time.ParseDuration(fmt.Sprintf("-%dm", rule.LeftEarly))
	if timeAttendance.After(timeOffworkBegin) && timeAttendance.Before(timeOffworkEnd) && rule.OffWorkCheck > 0 {
		return confreader.GetValue("Offwork"), strOffWorkTime, nil
	}

	if rule.LateTime > 0 && rule.LeftEarly > 0 {
		//严重迟到或早退
		if timeAttendance.After(timeWorkEnd.Add(lateSpan)) && timeAttendance.Before(timeOffworkBegin.Add(earlySpan)) {
			if rule.WorkCheck > 0 && rule.OffWorkCheck > 0 {
				if timeAttendance.Before(timeMidle) {
					return confreader.GetValue("SeriouslyLate"), strWorkTime, nil
				} else {
					return confreader.GetValue("SevereEarlyLeave"), strOffWorkTime, nil
				}
			}
		}
	}

	if timeAttendance.After(timeWorkEnd) && timeAttendance.Before(timeOffworkBegin) {
		if timeAttendance.Before(timeWorkEnd.Add(lateSpan)) && rule.WorkCheck > 0 {
			return confreader.GetValue("Late"), strWorkTime, nil
		}

		if timeAttendance.After(timeOffworkBegin.Add(earlySpan)) && rule.OffWorkCheck > 0 {
			return confreader.GetValue("EarlyLeave"), strOffWorkTime, nil
		}
	}

	return confreader.GetValue("NonAttendanceTimePeriod"), "", nil
}

func GetAttendanceShaduleResult(strTime string, nRuleID int64, nEmployeeID int64) (string, string, int64, error) {
	timeSplitArray := strings.Split(strTime, "T")
	if len(timeSplitArray) != 2 {
		return "", "", -1, errors.New("错误的创建时间")
	}
	timeSplitArray_ := strings.Split(timeSplitArray[1], "Z")
	strTime = timeSplitArray[0] + " " + timeSplitArray_[0]

	timeAttendance, err := time.Parse(TIME_LAYOUT, strTime)
	if err != nil {
		return "", "", -1, err
	}

	structRuleInfo, err := my_db.GetRuleByID(nRuleID)
	if err != nil {
		return "", "", -1, err
	}

	structEmployeeInfo, err := my_db.GetEmployeeById(nEmployeeID)
	if err != nil {
		return "", "", -1, err
	}

	arrayTags := strings.Split(structRuleInfo.Tags, ",")
	nIndex := -1
	for nTagIndex, strTagID := range arrayTags {
		nTagID, err := strconv.ParseInt(strTagID, 0, 64)
		if err != nil {
			return "", "", -1, err
		}
		if nTagID == structEmployeeInfo.TagID {
			nIndex = nTagIndex
			break
		}
	}

	if nIndex < 0 {
		return "", "", -1, errors.New("未找到相关班次")
	}

	arrayBegin := strings.Split(structRuleInfo.Begin, "T")
	arrayBegin_ := strings.Split(arrayBegin[1], "Z")
	strBeginTime := arrayBegin[0] + " " + arrayBegin_[0]
	timeBegin, err := time.Parse(TIME_LAYOUT, strBeginTime)
	if err != nil {
		return "", "", -1, err
	}

	timeSpan, err := time.ParseDuration("8h")
	if err != nil {
		return "", "", -1, err
	}
	timeBegin = timeBegin.Add(timeSpan)
	arrayBegin = strings.Split(timeBegin.Format(TIME_LAYOUT), " ")
	strBeginTime = arrayBegin[0] + " 00:00:00"
	timeBegin, err = time.Parse(TIME_LAYOUT, strBeginTime)
	if err != nil {
		return "", "", -1, err
	}

	timeDuration := timeAttendance.Sub(timeBegin)
	nDay := timeDuration.Hours() / 24

	if nDay < 0 {
		return "", "非考勤时段", -1, nil
	}

	for nDayIndex := 0; nDayIndex < int(nDay); nDayIndex++ {
		nIndex = (nIndex + 1) % len(arrayTags)
	}

	strTagID := arrayTags[nIndex]
	nTagID, err := strconv.ParseInt(strTagID, 0, 64)
	if err != nil {
		return "", "", -1, err
	}

	structTagInfo, err := my_db.GetTagByID(nTagID)
	if err != nil {
		return "", "", -1, err
	}

	strResult := ""

	if structTagInfo.Range1 == 0 && structTagInfo.Range2 == 0 && structTagInfo.Range3 == 0 {
		return "", "休息", -1, err
	}

	nRangeIndex := int64(0)
	strAttendanceTime := ""

	if structTagInfo.Range1 > 0 {
		nRangeIndex = 1
		structRangeInfo, err := my_db.GetRangeByID(structTagInfo.Range1)
		if err != nil {
			return "", "", -1, err
		}

		strResult, strAttendanceTime, err = GetAttendanceResultByRangeInfo(structRangeInfo, strTime)
		if err != nil {
			return "", "", -1, err
		}

	}

	if structTagInfo.Range2 > 0 && strResult == confreader.GetValue("NonAttendanceTimePeriod") {
		nRangeIndex = 2
		structRangeInfo, err := my_db.GetRangeByID(structTagInfo.Range2)
		if err != nil {
			return "", "", -1, err
		}

		strResult, strAttendanceTime, err = GetAttendanceResultByRangeInfo(structRangeInfo, strTime)
		if err != nil {
			return "", "", -1, err
		}

	}

	if structTagInfo.Range3 > 0 && strResult == confreader.GetValue("NonAttendanceTimePeriod") {
		nRangeIndex = 3
		structRangeInfo, err := my_db.GetRangeByID(structTagInfo.Range3)
		if err != nil {
			return "", "", -1, err
		}

		strResult, strAttendanceTime, err = GetAttendanceResultByRangeInfo(structRangeInfo, strTime)
		if err != nil {
			return "", "", -1, err
		}

	}

	if strResult == confreader.GetValue("Work") || strResult == confreader.GetValue("Offwork") || strResult == confreader.GetValue("SeriouslyLate") {
		return "", strResult, nRangeIndex, nil
	} else if strResult == confreader.GetValue("Offwork") || strResult == confreader.GetValue("EarlyLeave") || strResult == confreader.GetValue("SevereEarlyLeave") {
		return strResult, strAttendanceTime, nRangeIndex, nil
	} else {
		return strResult, strAttendanceTime, nRangeIndex, nil
	}

}

func GetAttendanceResultByChildRange(rule my_db.DayRangeInfo, strTime string) (string, string, error) {
	if rule.LateTime < 0 {
		rule.LateTime = 0
	}

	if rule.LeftEarly < 0 {
		rule.LeftEarly = 0
	}

	timeSplitArray := strings.Split(strTime, " ")
	if len(timeSplitArray) != 2 {
		return "", "", errors.New("错误的创建时间")
	}

	strWorkTime := rule.WorkBegin + "-" + rule.WorkEnd
	strOffworkTime := rule.OffworkBegin + "-" + rule.OffworkEnd

	rule.WorkBegin = timeSplitArray[0] + " " + rule.WorkBegin + ":00"
	rule.WorkEnd = timeSplitArray[0] + " " + rule.WorkEnd + ":00"
	rule.OffworkBegin = timeSplitArray[0] + " " + rule.OffworkBegin + ":00"
	rule.OffworkEnd = timeSplitArray[0] + " " + rule.OffworkEnd + ":00"

	//fmt.Println("时间数据:", rule.Work_begin, rule.Work_end, rule.Offwork_begin, rule.Offwork_end)

	timeWorkBegin, err := time.Parse(TIME_LAYOUT, rule.WorkBegin)
	if err != nil {
		return "", "", err
	}

	timeWorkEnd, err := time.Parse(TIME_LAYOUT, rule.WorkEnd)
	if err != nil {
		return "", "", err
	}

	timeOffworkBegin, err := time.Parse(TIME_LAYOUT, rule.OffworkBegin)
	if err != nil {
		return "", "", err
	}

	timeOffworkEnd, err := time.Parse(TIME_LAYOUT, rule.OffworkEnd)
	if err != nil {
		return "", "", err
	}

	timeSplitArray_ := strings.Split(timeSplitArray[1], "Z")

	strTime = timeSplitArray[0] + " " + timeSplitArray_[0]
	timeAttendance, err := time.Parse(TIME_LAYOUT, strTime)
	if err != nil {
		return "", "", err
	}

	bFind := false
	arrayDays := strings.Split(rule.Days, ",")
	for _, dayValue := range arrayDays {
		nDay, err := strconv.ParseInt(dayValue, 0, 64)
		if err != nil {
			return "", "", err
		}

		nWeekDay := int64(timeAttendance.Weekday())
		if nWeekDay == 0 {
			nWeekDay = 7
		}

		if nDay == nWeekDay {
			bFind = true
			break
		}
	}

	if !bFind {
		return confreader.GetValue("OffDay"), "", nil
	}

	//在上班打卡时间段   属于正常打卡
	lateSpan, err := time.ParseDuration(fmt.Sprintf("%dm", rule.LateTime))
	if timeAttendance.After(timeWorkBegin) && timeAttendance.Before(timeWorkEnd) && rule.WorkCheck > 0 {
		return confreader.GetValue("Work"), strWorkTime, nil
	}

	earlySpan, err := time.ParseDuration(fmt.Sprintf("-%dm", rule.LeftEarly))
	if timeAttendance.After(timeOffworkBegin) && timeAttendance.Before(timeOffworkEnd) && rule.WorkCheck > 0 {
		return confreader.GetValue("Offwork"), strOffworkTime, nil
	}

	if rule.LateTime > 0 && rule.LeftEarly > 0 {
		//严重迟到或早退
		if timeAttendance.After(timeWorkEnd.Add(lateSpan)) && timeAttendance.Before(timeOffworkBegin.Add(earlySpan)) {
			if rule.OffworkCheck > 0 && rule.WorkCheck > 0 {
				timeSpan := timeOffworkBegin.Sub(timeWorkBegin)
				timeSpan, err = time.ParseDuration(fmt.Sprintf("%fh", timeSpan.Hours()/2))
				if err != nil {
					return "", "", err
				}
				timeIndex := timeWorkBegin.Add(timeSpan)

				if timeAttendance.Before(timeIndex) {
					return confreader.GetValue("SeriouslyLate"), strWorkTime, nil
				} else {
					return confreader.GetValue("SevereEarlyLeave"), strOffworkTime, nil
				}
			}
		}
	}

	if timeAttendance.After(timeWorkEnd) && timeAttendance.Before(timeOffworkBegin) {
		if timeAttendance.Before(timeWorkEnd.Add(lateSpan)) && rule.WorkCheck > 0 {
			return confreader.GetValue("Late"), strWorkTime, nil
		}

		if timeAttendance.After(timeOffworkBegin.Add(earlySpan)) && rule.OffworkCheck > 0 {
			return confreader.GetValue("EarlyLeave"), strOffworkTime, nil
		}

	}

	return confreader.GetValue("NonAttendanceTimePeriod"), "", nil
}

//GetAttendnaceResultByDayRange __
func GetAttendnaceResultByDayRange(nFather int64, strTime string) (string, string, int64, error) {
	timeSplitArray := strings.Split(strTime, "T")
	if len(timeSplitArray) != 2 {
		return "", "", -1, errors.New("错误的创建时间:" + strTime)
	}
	timeSplitArray_ := strings.Split(timeSplitArray[1], "Z")
	strTime = timeSplitArray[0] + " " + timeSplitArray_[0]

	arrayRange, err := my_db.GetDayRangeByFather(nFather)
	if err != nil {
		beego.Error(err)
		return "", "", -1, err
	}

	nIndex := int64(0)

	strResult := ""
	strAttendanceTime := ""
	if len(arrayRange) > 0 {
		nIndex = 1
		strResult, strAttendanceTime, err = GetAttendanceResultByChildRange(arrayRange[0], strTime)
		if err != nil {
			return "", "", -1, err
		}
	}

	if len(arrayRange) > 1 && strResult == confreader.GetValue("NonAttendanceTimePeriod") {
		nIndex = 2
		strResult, strAttendanceTime, err = GetAttendanceResultByChildRange(arrayRange[1], strTime)
		if err != nil {
			return "", "", -1, err
		}
	}

	if len(arrayRange) > 2 && strResult == confreader.GetValue("NonAttendanceTimePeriod") {
		nIndex = 3
		strResult, strAttendanceTime, err = GetAttendanceResultByChildRange(arrayRange[2], strTime)
		if err != nil {
			return "", "", -1, err
		}
	}

	return strResult, strAttendanceTime, nIndex, nil
}
