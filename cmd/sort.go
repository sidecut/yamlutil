/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sidecut/yamlutil/argsprocessor"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var replace bool
var automaticName bool
var useStdIn bool

const eitherButNotBothErrorMessage = "--in or --file must be specified but not both"

// Usage:
// filenames are files to sort
// -r --replace -- do an in-place sort
// -a --auto -- automatic *.sorted.yaml filename
// If no filenames, use stdin

// sortCmd represents the sort command
var sortCmd = &cobra.Command{
	Use:   "sort",
	Short: "Sort YAML keys",
	Long: `Sorts YAML keys

Non-option arguments are names of files to sort.
If no filenames, use stdin.

-r --replace -- do an in-place sort
-a --auto -- automatic *.sorted.yaml filename`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 && (automaticName || replace) {
			err := errors.New("Can't use --auto or --replace with stdin")
			cobra.CheckErr(err)
		} else if len(args) != 0 && automaticName && replace {
			err := errors.New("Can't use both --auto and --replace")
			cobra.CheckErr(err)
		}

		argsprocessor.ProcessArgs(args, cmd.InOrStdin(), func(filename string, file io.Reader) {
			if len(args) > 1 && filename != argsprocessor.STDIN {
				cmd.OutOrStdout().Write([]byte(fmt.Sprintf("%v:\n", filename)))
			}

			if filename == argsprocessor.STDIN {
				err := doSortStdin(cmd)
				cobra.CheckErr(err)
			} else {
				outputFilename, err := getOutputFilename(filename)
				cobra.CheckErr(err)
				err = doSortFile(cmd, filename, outputFilename)
				cobra.CheckErr(err)
			}
		})
	},
}

func doSortStdin(cmd *cobra.Command) (err error) {
	yamlMap, err := getYamlMap(cmd.InOrStdin())
	if err != nil {
		return
	}

	err = writeSortedMap(yamlMap, cmd.OutOrStdout())
	if err != nil {
		return
	}

	return
}

func getOutputFilename(filename string) (string, error) {
	if replace {
		return filename, nil
	}
	if !automaticName {
		return argsprocessor.STDIN, nil
	}

	const out_yaml = ".sorted.yaml"
	parts := strings.Split(filename, ".")
	switch len(parts) {
	case 0:
		return argsprocessor.STDIN, errors.New("Empty filename")
	case 1:
		return filename + out_yaml, nil
	default:
		extension := parts[len(parts)-1]
		if strings.ToLower(extension) == "yaml" {
			return strings.Join(parts[:len(parts)-1], ".") + out_yaml, nil
		} else {
			return filename + out_yaml, nil
		}
	}
}

func doSortFile(cmd *cobra.Command, inputFilename string, outputFilename string) (err error) {
	input, err := os.Open(inputFilename)
	if err != nil {
		return
	}

	yamlMap, err := getYamlMap(input)
	if err != nil {
		return
	}
	err = input.Close()
	if err != nil {
		return
	}

	if outputFilename == argsprocessor.STDIN {
		output := cmd.OutOrStdout()
		err = writeSortedMap(yamlMap, output)
	} else {
		output, err := os.Create(outputFilename)
		if err != nil {
			return err
		}
		defer output.Close()

		err = writeSortedMap(yamlMap, output)
	}

	return
}

func getYamlMap(input io.Reader) (yamlMap map[string]interface{}, err error) {
	yamlBytes, err := ioutil.ReadAll(input)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(yamlBytes, &yamlMap)
	if err != nil {
		return
	}

	return
}

func writeSortedMap(yamlMap map[string]interface{}, output io.Writer) error {
	outBuffer, err := yaml.Marshal(yamlMap)
	if err != nil {
		return err
	}

	if _, err = output.Write(outBuffer); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(sortCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sortCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	sortCmd.Flags().BoolVarP(&replace, "replace", "r", false, "Do an in-place sort, replacing the file(s).")
	sortCmd.Flags().BoolVarP(&automaticName, "auto", "a", false, "Automatic *.sorted.yaml filename.")
}
