package config

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"time"
)

type Configuration struct {
	StartId     int    `json:"startid"`
	Userid      string `json:"userid"`
	From        string `json:"from"`
	Until       string `json:"until"`
	DailyBegin  string `json:"dailyBegin"`
	DailyEnd    string `json:"dailyEnd"`
	BeginDeltaS int32  `json:"beginDeltaS"`
	EndDeltaS   int32  `json:"endDeltaS"`
}

const (
	yearFormat string = "2006-01-02"
	hourFormat string = "15:04.05"
)

type ConfigurationError struct {
	Code int32
	Msg  string
}

func (cfg Configuration) ToJson() string {
	marshal, _ := json.Marshal(cfg)
	return string(marshal)
}
func (r *ConfigurationError) Error() string {
	return fmt.Sprintf("Error code %d: err %v", r.Code, r.Msg)
}

func (cfg *Configuration) FromDate() time.Time {
	loc, _ := time.LoadLocation("Local")
	parse, _ := time.ParseInLocation(yearFormat, cfg.From, loc)
	return parse.UTC()

}

func (cfg *Configuration) UntilDate() time.Time {
	loc, _ := time.LoadLocation("Local")
	parse, _ := time.ParseInLocation(yearFormat, cfg.Until, loc)
	return parse.UTC()

}

func (cfg *Configuration) DailyBeginHour() time.Time {
	loc, _ := time.LoadLocation("Local")
	parse, _ := time.ParseInLocation(hourFormat, cfg.DailyBegin, loc)
	return parse.UTC()

}

func (cfg *Configuration) DailyEndHour() time.Time {
	loc, _ := time.LoadLocation("Local")
	parse, _ := time.ParseInLocation(hourFormat, cfg.DailyEnd, loc)
	return parse.UTC()

}
func (cfg *Configuration) validate() error {

	if cfg.StartId < 0 {
		return &ConfigurationError{
			Code: 4,
			Msg:  fmt.Sprintf("negative id not allowed %d", cfg.StartId),
		}
	}
	{
		_, err := time.Parse(yearFormat, cfg.From)
		if err != nil {
			return &ConfigurationError{
				Code: 3,
				Msg:  fmt.Sprintf("Can't parse From date %s", err.Error()),
			}
		}
	}

	{
		_, err := time.Parse(yearFormat, cfg.Until)
		if err != nil {
			return &ConfigurationError{
				Code: 3,
				Msg:  fmt.Sprintf("Can't parse Until date %s", err.Error()),
			}
		}
	}

	{
		_, err := time.Parse(hourFormat, cfg.DailyBegin)
		if err != nil {
			return &ConfigurationError{
				Code: 3,
				Msg:  fmt.Sprintf("Can't parse DailyBegin date %s", err.Error()),
			}
		}
	}

	{
		_, err := time.Parse(hourFormat, cfg.DailyEnd)
		if err != nil {
			return &ConfigurationError{
				Code: 3,
				Msg:  fmt.Sprintf("Can't parse DailyEnd date %s", err.Error()),
			}
		}
	}
	if cfg.FromDate().After(cfg.UntilDate()) {
		return &ConfigurationError{
			Code: 0,
			Msg:  fmt.Sprintf("From after Until. %v is can't be after %v", cfg.FromDate(), cfg.UntilDate()),
		}
	}

	if cfg.DailyBeginHour().After(cfg.DailyEndHour()) {
		return &ConfigurationError{
			Code: 1,
			Msg:  fmt.Sprintf("Daily begin after daily end. %v is can't be after %v", cfg.DailyBeginHour(), cfg.DailyEndHour()),
		}
	}

	return nil
}

func (cfg *Configuration) Generate(writer io.Writer) error {
	err := cfg.validate()
	if err != nil {
		return err
	}
	sub := int(cfg.UntilDate().Sub(cfg.FromDate()).Hours() / 24)
	dayCounter := cfg.FromDate().Add(time.Hour * -24)
	insertStatement := "INSERT INTO `user_times` (`id`, `crdate`, `cruser_id`, `modified`, `user_id`, `date`, `starttime`, `endtime`, `calctime`, `client`, `project`, `task`, `description`, `time_type_id`, `deducted`, `clearable`, `disabled`, `deleted`) VALUES"

	_, err = writer.Write([]byte(insertStatement))
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte("\n"))
	if err != nil {
		return err
	}
	for i := 0; i <= sub; i++ {
		rowId := cfg.StartId + i
		dayCounter = dayCounter.Add(time.Hour * 24)
		if dayCounter.Weekday() == time.Saturday ||
			dayCounter.Weekday() == time.Sunday {
			continue
		}
		beginOffset := time.Duration(rand.Int31n(cfg.BeginDeltaS))
		endOffset := time.Duration(rand.Int31n(cfg.EndDeltaS))
		fromHour := cfg.DailyBeginHour().Add(time.Second * beginOffset)
		untilHour := cfg.DailyEndHour().Add(time.Second * endOffset)

		//fromDate := toTime(dayCounter, fromHour)
		//untilDate := toTime(dayCounter, untilHour)
		//writer.Write([]byte(fmt.Sprintf("Day: %v: from: %v until:%v\n",
		//	dayCounter.Format("2006-01-02 Monday -0700"),
		//	fromDate.Format("2006-01-02 Monday 15:04.05 -0700 "),
		//	untilDate.Format("2006-01-02 Monday 15:04.05 -0700"))))

		fromDate := toTime(dayCounter, fromHour)
		untilDate := toTime(dayCounter, untilHour)
		//writer.Write([]byte(fmt.Sprintf("Day: %v: from: %v until:%v\n",
		//	dayCounter.Format("2006-01-02 Monday"),
		//	fromDate.Unix(),
		//	untilDate.Unix())))
		statement := valuesStatement(rowId, cfg.Userid, fromDate.Unix(), untilDate.Unix(), dayCounter.Unix(), i == sub)
		_, err := writer.Write([]byte(fmt.Sprintf("%s\n", statement)))
		if err != nil {
			return err
		}
	}

	return nil
}

func valuesStatement(rowId int, userid string, fromEpoch int64, untilEpoch int64, epochCurrentDay int64, lastStatement bool) string {

	sub := time.Unix(untilEpoch, 0).Sub(time.Unix(fromEpoch, 0))

	worktime := fmt.Sprintf("%s", fmtDuration(sub))
	//time.Date(2023, 01, 01, sub.Hours(), sub.Minutes(), 0, 0, time.UTC)
	sprintf := fmt.Sprintf("(%d, %d, 0, %d, %s, %d, %d, %d, '%s', 470, 0, 5, 'Arbeitszeit', 0, 0, 1, 0, 0)",
		rowId,           // ID
		fromEpoch,       // crdate
		untilEpoch,      // modified
		userid,          //user_id
		epochCurrentDay, //date day of the month
		fromEpoch,       // Starttime
		untilEpoch,      // endtime
		worktime,
	)

	if lastStatement {
		return sprintf + ";"
	} else {
		return sprintf + ","
	}
}
func toTime(theDay time.Time, theHour time.Time) time.Time {
	date := time.Date(theDay.Year(), theDay.Month(), theDay.Day(), theHour.Hour(), theHour.Minute(), theHour.Second(), theHour.Nanosecond(), time.UTC)
	return date
}

func fmtDuration(duration time.Duration) string {
	duration = duration.Round(time.Minute)
	h := duration / time.Hour
	duration -= h * time.Hour
	m := duration / time.Minute
	return fmt.Sprintf("%02d:%02d", h, m)
}
