package main

import (
	"bufio"
	"flag"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/suikast42/ramostamp/config"
	"os"
)

func main() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("Start reading configuration")

	err := readConfig()
	if err != nil {
		panic(err)
	}
	//parseYear, _ := time.Parse("2006-01-02", "2042-11-12")
	//parseHours, _ := time.Parse("15:04.05", "23:17.42")
	//log.Info().Msgf("%v", parseYear)
	//log.Info().Msgf("%v", parseHours)
	var configuration config.Configuration
	err = viper.Unmarshal(&configuration)
	if err != nil {
		panic(err)
	}

	log.Info().Msgf("Configuration:\n%v", configuration.ToJson())

	var out = flag.String("out", "stdout", "Write the generated output to file instead of stdout")
	withComment := flag.Bool("comment", false, "Generate with comment")
	flag.Parse()

	if "stdout" == *out {
		err = configuration.Generate(os.Stderr, *withComment)
	} else {
		f, err := os.Create(*out)
		if err != nil {
			log.Err(err).Msgf("Can't create file %s", *out)
			os.Exit(1)
		}
		writer := bufio.NewWriter(f)
		configuration.Generate(writer, *withComment)
		writer.Flush()
		log.Info().Msgf("Output wrote to %s", *out)
	}
	if err != nil {
		log.Error().Msgf("Error in generation process:\n%s", err.Error())
	}

}

func readConfig() error {
	viper.SetConfigType("json") // Look for specific type
	{

		//initialize local cfg
		viper.AddConfigPath("./")
		viper.SetConfigName("config") // Register config file name (no extension)
		err := viper.ReadInConfig()
		if err != nil {
			return err
		}
	}

	return nil
}
