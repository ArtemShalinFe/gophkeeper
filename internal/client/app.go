package client

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

type App struct {
	gkclient *GAgentClient
	login    string
	password string
}

func NewApp() *App {
	return &App{
		gkclient: NewGAgentClient(),
	}
}

const (
	usernameFullFlagName     = "username"
	usernameShortFlagName    = "u"
	passwordFullFlagName     = "password"
	passwordShortFlagName    = "p"
	descriptionFullFlagName  = "description"
	descriptionShortFlagName = "d"
	metainfoFullFlagName     = "metainfo"
	metainfoShortFlagName    = "m"
	recordIDFullFlagName     = "recordID"
	recordIDShortFlagName    = "r"
)

func (a *App) Execute() {
	var rootCmd = &cobra.Command{
		Use:   "gkeeper",
		Short: "gkeeper is client for gophkeeper application",
		Long: `gkeeper is client for gophkeeper application.
		
		gophkeeper is an application that securely stores: 
		- usernames
		- passwords
		- binary data and other personal information`,
	}

	a.initUsersCmd(rootCmd)
	a.initRecordsCmd(rootCmd)

	rootCmd.PersistentFlags().StringVarP(&a.gkclient.Addr, "address", "a", "", "Gophkeeper agent service address")
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("execute command, err: %v", err)
	}
}

func (a *App) initUsersCmd(rootCmd *cobra.Command) {
	var usersCmd = &cobra.Command{
		Use:   "users",
		Short: "User management",
		Long:  `Allows to register in the gophkeeper service and receive a token to work with it`,
	}

	rootCmd.AddCommand(usersCmd)
	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "User registration",
		Long:  `User registration in the gophkeeper service`,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := a.gkclient.Register(rootCmd.Context(), &models.UserDTO{
				Login:    a.login,
				Password: a.password,
			})
			if err != nil {
				return fmt.Errorf("an error occured while register, err: %w", err)
			}

			fmt.Print(token)

			return nil
		},
	}
	registerCmd.Flags().StringVarP(&a.login, usernameFullFlagName, usernameShortFlagName, "",
		"register login")
	registerCmd.Flags().StringVarP(&a.password, passwordFullFlagName, passwordShortFlagName, "",
		"register password")

	usersCmd.AddCommand(registerCmd)

	usersCmd.Flags().StringVarP(&a.login, usernameFullFlagName, usernameShortFlagName, "",
		"log in with this login")
	usersCmd.Flags().StringVarP(&a.password, passwordFullFlagName, passwordShortFlagName, "",
		"log in with this password")

	usersCmd.AddCommand(&cobra.Command{
		Use:   "login",
		Short: "Working with the token",
		Long:  `Getting a token to perform storage operations`,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := a.gkclient.Login(rootCmd.Context(), &models.UserDTO{
				Login:    a.login,
				Password: a.password,
			})
			if err != nil {
				return fmt.Errorf("an error occured while login, err: %w", err)
			}

			fmt.Print(token)

			return nil
		},
	})
}

func (a *App) initRecordsCmd(rootCmd *cobra.Command) {
	var recordsCmd = &cobra.Command{
		Use:   "records",
		Short: "Managing records",
		Long:  `Command palette for managing records in the agent storage`,
	}

	a.initListCmd(recordsCmd)
	a.initAddCmd(recordsCmd)
	a.initGetCmd(recordsCmd)
	a.initUpdateCmd(recordsCmd)
	a.initDeleteCmd(recordsCmd)

	recordsCmd.Flags().StringVarP(&a.gkclient.token, "token", "t", "", "JWT for gophkeeper agent service address")

	rootCmd.AddCommand(recordsCmd)
}

func (a *App) initListCmd(recordsCmd *cobra.Command) {
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "Return user records",
		Long:  `Returns a list of objects available to the user`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := a.gkclient.GetRecords(recordsCmd.Context())
			if err != nil {
				return fmt.Errorf("an error occured while retrieving list of records, err: %w", err)
			}

			fmt.Print(models.RecordStringHeader())
			for _, r := range rs {
				fmt.Println(r)
			}

			return nil
		},
	}

	recordsCmd.AddCommand(listCmd)
}

func (a *App) initAddCmd(recordsCmd *cobra.Command) {
	type commonVars struct {
		description string
		metainfo    []string
	}
	var cmdVars commonVars

	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Adds an entry to the agent",
		Long:  `Adds an entry to the agent cache for subsequent synchronization`,
	}

	a.initAddAuthCmd(addCmd)
	a.initAddTextCmd(addCmd)
	a.initAddBinaryCmd(addCmd)
	a.initAddCardCmd(addCmd)

	recordsCmd.AddCommand(addCmd)
	recordsCmd.PersistentFlags().StringVarP(&cmdVars.description, descriptionFullFlagName, descriptionShortFlagName, "",
		"This description will be added in record")
	recordsCmd.PersistentFlags().StringArrayVarP(&cmdVars.metainfo, metainfoFullFlagName, metainfoShortFlagName,
		make([]string, 0),
		`Any textual meta information containing a '<key>:<value>' pair separated by commas, that will be added. 
	Example: "path:sompath,keyword:somekeyword"`)
}

func (a *App) initUpdateCmd(recordsCmd *cobra.Command) {
	type commonVars struct {
		recordID    string
		description string
		metainfo    []string
	}
	var cmdVars commonVars
	var updateCmd = &cobra.Command{
		Use:   "add",
		Short: "Adds an entry to the agent",
		Long:  `Adds an entry to the agent cache for subsequent synchronization`,
	}

	updateCmd.Flags().StringVarP(&cmdVars.recordID, recordIDFullFlagName, recordIDShortFlagName, "",
		"Record id that will be updated")
	updateCmd.PersistentFlags().StringVarP(&cmdVars.description, descriptionFullFlagName, descriptionShortFlagName, "",
		"This description will be updated in record")
	updateCmd.PersistentFlags().StringArrayVarP(&cmdVars.metainfo, metainfoFullFlagName, metainfoShortFlagName,
		make([]string, 0),
		`Any textual meta information containing a '<key>:<value>' pair separated by commas, that will be updated. 
	Example: "path:sompath,keyword:somekeyword"`)

	a.initAddAuthCmd(updateCmd)
	a.initAddTextCmd(updateCmd)
	a.initAddBinaryCmd(updateCmd)
	a.initAddCardCmd(updateCmd)

	recordsCmd.AddCommand(updateCmd)
}

var tmpParseMetainfoErr = "an error occured while parse metainfo, err: %w"

func (a *App) initAddAuthCmd(addCmd *cobra.Command) {
	type authVars struct {
		username    string
		keyword     string
		description string
		metainfo    []string
	}
	var cmdVars authVars

	var authCmd = &cobra.Command{
		Use:   "auth",
		Short: "login and password",
		Long:  `The data type that contains the logs and password that will be stored in the storage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mi, err := models.NewMetainfoFromStringArray(cmdVars.metainfo)
			if err != nil {
				return fmt.Errorf(tmpParseMetainfoErr, err)
			}
			rdto, err := models.NewRecordDTO(
				cmdVars.description,
				models.AuthType,
				&models.Auth{
					Login:    cmdVars.username,
					Password: cmdVars.keyword,
				},
				mi,
			)
			if err != nil {
				return fmt.Errorf("an error occured while created auth record dto, err: %w", err)
			}

			r, err := a.gkclient.AddRecord(addCmd.Context(), rdto)
			if err != nil {
				return fmt.Errorf("an error occured while add auth record, err: %w", err)
			}

			fmt.Print(r.ID)

			return nil
		},
	}

	authCmd.Flags().StringVarP(&cmdVars.username, usernameFullFlagName, usernameShortFlagName, "",
		"stored username")
	authCmd.Flags().StringVarP(&cmdVars.keyword, passwordFullFlagName, passwordShortFlagName, "",
		"stored keyword or password")

	addCmd.AddCommand(authCmd)
}

func (a *App) initAddTextCmd(addCmd *cobra.Command) {
	type textVars struct {
		text        string
		description string
		metainfo    []string
	}
	var cmdVars textVars

	var textCmd = &cobra.Command{
		Use:   "text",
		Short: "Any text",
		Long:  `Any text`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mi, err := models.NewMetainfoFromStringArray(cmdVars.metainfo)
			if err != nil {
				return fmt.Errorf(tmpParseMetainfoErr, err)
			}
			rdto, err := models.NewRecordDTO(
				cmdVars.description,
				models.TextType,
				&models.Text{
					Data: cmdVars.text,
				},
				mi,
			)
			if err != nil {
				return fmt.Errorf("an error occured while created text record dto, err: %w", err)
			}

			r, err := a.gkclient.AddRecord(addCmd.Context(), rdto)
			if err != nil {
				return fmt.Errorf("an error occured while add text record, err: %w", err)
			}

			fmt.Print(r.ID)

			return nil
		},
	}

	textCmd.Flags().StringVar(&cmdVars.text, "text", "", "stored text")

	addCmd.AddCommand(textCmd)
}

func (a *App) initAddBinaryCmd(addCmd *cobra.Command) {
	type binaryVars struct {
		path        string
		description string
		metainfo    []string
	}
	var cmdVars binaryVars

	var binaryCmd = &cobra.Command{
		Use:   "binary",
		Short: "Any files. Limited to 20 MB in size",
		Long:  `Any files. Limited to 20 MB in size.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			i, err := os.Stat(cmdVars.path)
			if err != nil {
				return fmt.Errorf("an error occured while retrieving file info, err: %w", err)
			}
			if i.Size() > int64(models.MaxFileSize) {
				return errors.New("file too large")
			}

			d, err := os.ReadFile(cmdVars.path)
			if err != nil {
				return fmt.Errorf("an error occured while read file, err: %w", err)
			}

			mi, err := models.NewMetainfoFromStringArray(cmdVars.metainfo)
			if err != nil {
				return fmt.Errorf(tmpParseMetainfoErr, err)
			}

			rdto, err := models.NewRecordDTO(
				cmdVars.description,
				models.BinaryType,
				&models.Binary{
					Data: d,
				},
				mi,
			)
			if err != nil {
				return fmt.Errorf("an error occured while created binary record dto, err: %w", err)
			}

			r, err := a.gkclient.AddRecord(addCmd.Context(), rdto)
			if err != nil {
				return fmt.Errorf("an error occured while add binary record, err: %w", err)
			}

			fmt.Print(r.ID)

			return nil
		},
	}

	binaryCmd.Flags().StringVarP(&cmdVars.path, "path", "p", "", "path to file")

	addCmd.AddCommand(binaryCmd)
}

func (a *App) initAddCardCmd(addCmd *cobra.Command) {
	type cardVars struct {
		number      string
		term        string
		owner       string
		description string
		metainfo    []string
	}
	var cmdVars cardVars

	var binaryCmd = &cobra.Command{
		Use:   "card",
		Short: "The bank card details",
		Long:  `The bank card details including: number, term and owner. cvv code is not stored!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := time.Parse("MM/YY", cmdVars.term)
			if err != nil {
				return fmt.Errorf("an error occured while parse time, err: %w", err)
			}

			mi, err := models.NewMetainfoFromStringArray(cmdVars.metainfo)
			if err != nil {
				return fmt.Errorf(tmpParseMetainfoErr, err)
			}

			rdto, err := models.NewRecordDTO(
				cmdVars.description,
				models.BinaryType,
				&models.Card{
					Number: cmdVars.number,
					Term:   t,
					Owner:  cmdVars.owner,
				},
				mi,
			)
			if err != nil {
				return fmt.Errorf("an error occured while created record dto, err: %w", err)
			}

			r, err := a.gkclient.AddRecord(addCmd.Context(), rdto)
			if err != nil {
				return fmt.Errorf("an error occured while add card record, err: %w", err)
			}

			fmt.Print(r.ID)

			return nil
		},
	}

	binaryCmd.Flags().StringVar(&cmdVars.number, "number", "", "card number")
	binaryCmd.Flags().StringVar(&cmdVars.term, "term", "", "card term")
	binaryCmd.Flags().StringVar(&cmdVars.owner, "owner", "", "card owner")

	addCmd.AddCommand(binaryCmd)
}

func (a *App) initDeleteCmd(recordsCmd *cobra.Command) {
	type cardVars struct {
		recordID string
	}
	var cmdVars cardVars
	var delCmd = &cobra.Command{
		Use:   "delete",
		Short: "Deletes an entry to the agent",
		Long:  `Mark an entry for removing from the agent cache for subsequent synchronization`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.gkclient.DeleteRecord(recordsCmd.Context(), cmdVars.recordID); err != nil {
				return fmt.Errorf("an error occured while mark record for removing, err: %w", err)
			}

			fmt.Print("done")

			return nil
		},
	}

	delCmd.Flags().StringVarP(&cmdVars.recordID, recordIDFullFlagName, recordIDShortFlagName, "",
		"Record id that will be mark as deleted")

	recordsCmd.AddCommand(delCmd)
}

func (a *App) initGetCmd(recordsCmd *cobra.Command) {
	var recordID string
	var getCmd = &cobra.Command{
		Use:   "delete",
		Short: "Deletes an entry to the agent",
		Long:  `Mark an entry for removing from the agent cache for subsequent synchronization`,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := a.gkclient.GetRecord(recordsCmd.Context(), recordID)
			if err != nil {
				return fmt.Errorf("an error occured get record from storage, err: %w", err)
			}

			fmt.Print(models.RecordStringHeader())
			fmt.Println(r)

			return nil
		},
	}

	getCmd.Flags().StringVarP(&recordID, recordIDFullFlagName, recordIDShortFlagName, "",
		"Record id that will retrieving from agent")

	recordsCmd.AddCommand(getCmd)
}
