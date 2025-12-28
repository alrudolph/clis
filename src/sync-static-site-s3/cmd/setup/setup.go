package setup

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/alrudolph/clis/src/sync-static-site-s3/cmd"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Save directory and profile for future use",
	// 	Long: `This CLI uploads the files in a directory to an S3 bucket but makes important
	// considerations for html files. If the file is an html file, the .html extension is removed
	// except for index.html and error.html, since we no longer have the html extension we also need
	// to set the file's Content-Type. This is necessary to allow the files to be accessed without
	// the .html extension, for example, domain.com/file instead of domain.com/file.html.

	// Example Usage:
	// 	go run . --directory /path/to/static/site --bucket s3-bucket-name
	// `,
	Run: func(command *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Println("Additional supplied args will be ignored")
		}

		config, err := cmd.NewConfig(command, args)

		if err != nil {
			log.Fatal(err)
		}

		configName, _ := command.Flags().GetString("config-name")
		userDirectory, err := filepath.Abs(".")

		if err != nil {
			log.Fatalf("Error getting relative path: %v", err)
		}

		toSave := cmd.SavedConfig{
			UserDirectory:   userDirectory,
			Name:            configName,
			Region:          config.Region,
			AccessKeyID:     config.AccessKeyID,
			SecretAccessKey: config.SecretAccessKey,
			Profile:         config.Profile,
			Role:            config.Role,
			Bucket:          config.Bucket,
			Directory:       config.Directory,
		}

		file := cmd.SavedConfigFile{}

		usr, err := user.Current()

		if err != nil {
			log.Fatal(err)
			return
		}

		homeDir := usr.HomeDir

		// if file doesn't exist, create
		fileName := filepath.Join(homeDir, "sync-s3", "config.json")

		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			if os.MkdirAll(filepath.Join(homeDir, "sync-s3"), os.ModePerm) != nil {
				log.Fatal(err)
				return
			}

			newFile, err := os.Create(fileName)
			if err != nil {
				log.Fatal(err)
				newFile.Close()
				return
			}
			newFile.Close()

			err = os.WriteFile(fileName, []byte(`{"profiles": []}`), 0644)
			if err != nil {
				log.Fatal(err)
				return
			}
		}

		// Read from file
		configFile, err := os.Open(fileName)

		if err != nil {
			log.Fatal(err)
			configFile.Close()
			return
		}

		savedConfig := cmd.SavedConfigFile{}

		// load json into struct
		jsonParser := json.NewDecoder(configFile)
		if err = jsonParser.Decode(&savedConfig); err != nil {
			log.Fatal(err)
			configFile.Close()
			return
		}
		configFile.Close()

		// already have this config, ignore
		for _, profile := range savedConfig.Profiles {
			if profile.Name == toSave.Name && profile.UserDirectory == toSave.UserDirectory {
				log.Fatal("Profile already exists")
				return
			}
		}

		// add new data
		file.Profiles = append(file.Profiles, toSave)

		// write struct to file
		writeToFile, err := json.Marshal(file)

		if err != nil {
			log.Fatal(err)
			return
		}

		err = os.WriteFile(fileName, writeToFile, 0644)
		if err != nil {
			log.Fatal(err)
			return
		}
	},
}

func init() {
	setupCmd.Flags().StringP("config-name", "", "", "Save multiple configs for this directory")
	_ = setupCmd.MarkFlagRequired("config-name")

	setupCmd.Flags().StringP("directory", "d", "", "Path to the static site directory")
	_ = setupCmd.MarkFlagDirname("directory")
	_ = setupCmd.MarkFlagRequired("directory")

	setupCmd.Flags().StringP("bucket", "b", "", "S3 bucket name")
	_ = setupCmd.MarkFlagRequired("bucket")

	setupCmd.Flags().StringP("region", "r", "us-east-1", "S3 bucket region")

	setupCmd.Flags().String("access-key-id", "", "AWS Access Key ID")
	setupCmd.Flags().String("secret-access-key", "", "AWS Secret Access Key")
	setupCmd.Flags().StringP("profile", "p", "", "AWS Profile name")
	setupCmd.Flags().StringP("role", "", "", "Role to switch into")

	cmd.RootCmd.AddCommand(setupCmd)
}
