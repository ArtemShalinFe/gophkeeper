package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
	"github.com/fxamacker/cbor/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func addTableHeaderCell(name string) *tview.TableCell {
	return tview.NewTableCell(name).
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignCenter)
}

func addTableCell(name string) *tview.TableCell {
	return tview.NewTableCell(name).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft)
}

const (
	pageListRecords        = "List records"
	pageAddAuthRecord      = "New auth record"
	pageUpdateAuthRecord   = "Update auth record"
	pageAddTextRecord      = "New text record"
	pageUpdateTextRecord   = "Update text record"
	pageAddBinaryRecord    = "New binary record"
	pageUpdateBinaryRecord = "Update binary record"
	pageAddCardRecord      = "New card record"
	pageUpdateCardRecord   = "Update card record"
	loginPage              = "Login page"
)

const (
	fnDescription          = "Description"
	fnUsername             = "Username"
	fnPassword             = "Password"
	fnMetadata             = "Metadata"
	fnPath                 = "Path"
	fnText                 = "Text"
	fnNumber               = "Number"
	fnOwner                = "Owner"
	fnTerm                 = "Term"
	fnTemplateTermDesc     = "Template for term"
	fnTemplateHintTermDesc = "Please enter Term in format MM/YY, where MM - month, YY - year"
	fnDateFormat           = "02/01/2006 03:04.000"
)

const (
	colID = iota
	colDesc
	colCreated
	colModified
	colType
	colHash
	colVersion
)

const defFileMode = 0600

func (ui *TUI) displayRecords(ctx context.Context, offset int, limit int) {
	curOst := offset
	curLt := limit

	rs, err := ui.authUser.GetRecords(ctx, ui.cache, curOst, curLt)
	if err != nil {
		ui.displayErr(fmt.Sprintf("an error occured while retrieving record list, err: %v", err))
		return
	}

	table := tview.NewTable()

	table.SetCell(0, colID, addTableHeaderCell("ID"))
	table.SetCell(0, colDesc, addTableHeaderCell(strings.ToUpper(fnDescription)))
	table.SetCell(0, colCreated, addTableHeaderCell("CREATED"))
	table.SetCell(0, colModified, addTableHeaderCell("MODIFIED"))
	table.SetCell(0, colType, addTableHeaderCell("TYPE"))
	table.SetCell(0, colHash, addTableHeaderCell("HASHSUM"))
	table.SetCell(0, colVersion, addTableHeaderCell("VERSION"))

	for r := curOst; r < len(rs); r++ {
		record := rs[r]
		rn := r + 1

		table.SetCell(rn, colID, addTableCell(record.ID))
		table.SetCell(rn, colDesc, addTableHeaderCell(record.Description))
		table.SetCell(rn, colCreated, addTableHeaderCell(record.Created.Format(fnDateFormat)))
		table.SetCell(rn, colModified, addTableHeaderCell(record.Modified.Format(fnDateFormat)))
		table.SetCell(rn, colType, addTableHeaderCell(record.Type))
		table.SetCell(rn, colHash, addTableHeaderCell(record.Hashsum))
		table.SetCell(rn, colVersion, addTableHeaderCell(strconv.FormatInt(record.GetVersion(), 10)))

		if rn >= curOst+ui.recLimit {
			break
		}
	}
	table.SetSelectable(true, false)

	table.SetSelectedFunc(func(row int, column int) {
		recordID := table.GetCell(row, colID).Text
		if strings.TrimSpace(recordID) == "" {
			ui.displayErr("record id is empty")
			return
		}

		dataType := table.GetCell(row, colType).Text
		switch dataType {
		case string(models.AuthType):
			ui.displayUpdateAuth(ctx, recordID)
		case string(models.TextType):
			ui.displayUpdateText(ctx, recordID)
		case string(models.BinaryType):
			ui.displayUpdateBinary(ctx, recordID)
		case string(models.CardType):
			ui.displayUpdateCard(ctx, recordID)
		default:
			ui.displayErr("Unknow type")
		}
	})

	buttonsManageList := tview.NewForm().
		AddButton("<", func() {
			curOst := max(0, curOst-curLt)

			ui.pages.RemovePage(pageListRecords)
			ui.displayRecords(ctx, curOst, curLt)
		}).
		AddButton("Refresh", func() {
			ui.pages.RemovePage(pageListRecords)
			ui.displayRecords(ctx, curOst, curLt)
		}).
		AddButton(">", func() {
			curOst := curOst + curLt

			ui.pages.RemovePage(pageListRecords)
			ui.displayRecords(ctx, curOst, curLt)
		}).
		AddButton("Back to menu", func() {
			ui.pages.RemovePage(pageListRecords)
		})
	buttonsManageList.SetButtonsAlign(tview.AlignLeft).
		SetBorderPadding(0, 0, 0, 0)

	buttons := tview.NewForm().
		AddButton("Add auth", func() { ui.displayCreateAuth(ctx) }).
		AddButton("Add text", func() { ui.displayCreateText(ctx) }).
		AddButton("Add file", func() { ui.displayCreateBinary(ctx) }).
		AddButton("Add card", func() { ui.displayCreateCard(ctx) })

	buttons.SetButtonsAlign(tview.AlignLeft).SetBorderPadding(0, 0, 0, 0)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(buttons, 1, 1, false).
		AddItem(table, 0, 1, true).
		AddItem(buttonsManageList, 1, 1, false)

	flex.SetBorder(true)

	ui.pages.AddPage(pageListRecords, flex, true, true)
}

func (ui *TUI) displayCreateAuth(ctx context.Context) {
	var desc string
	var login string
	var pass string
	var metadata []*models.Metadata
	mit := ""
	form := tview.NewForm().
		AddInputField(fnDescription, "", defaultFieldWidth, nil, func(v string) {
			desc = v
		}).
		AddInputField(fnUsername, "", defaultFieldWidth, nil, func(v string) {
			login = v
		}).
		AddInputField(fnPassword, "", defaultFieldWidth, nil, func(v string) {
			pass = v
		}).
		AddTextArea(fnMetadata, mit, defaultFieldWidth, 0, 0, func(text string) {
			mit = text
		})

	form.SetTitle(pageAddAuthRecord).
		SetTitleAlign(tview.AlignLeft)

	buttons := tview.NewForm().
		AddButton(buttonOkDesc, func() {
			auth := &models.Auth{
				Login:    login,
				Password: pass,
			}
			rdto, err := models.NewRecordDTO(
				desc,
				models.AuthType,
				auth,
				metadata,
			)

			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			rows := splitMetadata(mit)
			m, err := models.NewMetadataFromStringArray(rows)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			rdto.Metadata = m

			_, err = ui.authUser.AddRecord(ctx, ui.cache, rdto)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			ui.pages.RemovePage(pageAddAuthRecord)
		}).
		AddButton(buttonCancelDesc, func() { ui.pages.RemovePage(pageAddAuthRecord) })

	buttons.SetButtonsAlign(tview.AlignLeft).SetBorderPadding(0, 0, 0, 0)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(buttons, 1, 1, false)

	ui.pages.AddPage(pageAddAuthRecord, flex, true, true)
}

func (ui *TUI) displayUpdateAuth(ctx context.Context, recordID string) {
	r, err := ui.authUser.GetRecord(ctx, ui.cache, recordID)
	if err != nil {
		ui.displayErr(err.Error())
	}

	a := &models.Auth{}
	if err := cbor.Unmarshal(r.Data, a); err != nil {
		ui.displayErr(fmt.Sprintf("unmarshal data auth binary, err: %v", err))
		return
	}

	desc := r.Description
	login := a.Login
	pass := a.Password
	metadata := r.Metadata
	mit := convertMetadataToString(r.Metadata)
	form := tview.NewForm().
		AddInputField(fnDescription, r.Description, defaultFieldWidth, nil, func(v string) {
			desc = v
		}).
		AddInputField(fnUsername, a.Login, defaultFieldWidth, nil, func(v string) {
			login = v
		}).
		AddInputField(fnPassword, a.Password, defaultFieldWidth, nil, func(v string) {
			pass = v
		}).
		AddTextArea(fnMetadata, mit, defaultFieldWidth, 0, 0, func(text string) {
			mit = text
		})

	form.SetTitle(pageUpdateAuthRecord).
		SetTitleAlign(tview.AlignLeft)

	buttons := tview.NewForm().
		AddButton(buttonUpdate, func() {
			auth := &models.Auth{
				Login:    login,
				Password: pass,
			}

			r, err := models.NewRecord(
				r.ID,
				desc,
				models.AuthType,
				r.Created,
				time.Now(),
				auth,
				metadata,
				false,
				r.Version,
			)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			rows := splitMetadata(mit)
			m, err := models.NewMetadataFromStringArray(rows)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			r.Metadata = m

			r.Version++
			_, err = ui.authUser.UpdateRecord(ctx, ui.cache, r)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			ui.pages.RemovePage(pageUpdateAuthRecord)
		}).
		AddButton(buttonCancelDesc, func() { ui.pages.RemovePage(pageUpdateAuthRecord) })

	buttons.SetButtonsAlign(tview.AlignLeft).SetBorderPadding(0, 0, 0, 0)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).AddItem(buttons, 1, 1, false)

	ui.pages.AddPage(pageUpdateAuthRecord, flex, true, true)
}

func (ui *TUI) displayCreateText(ctx context.Context) {
	var text string
	var desc string
	var metadata []*models.Metadata
	mit := ""
	form := tview.NewForm().
		AddInputField(fnDescription, "", defaultFieldWidth, nil, func(v string) {
			desc = v
		}).
		AddInputField(fnText, "", defaultFieldWidth, nil, func(v string) {
			text = v
		}).
		AddTextArea(fnMetadata, mit, defaultFieldWidth, 0, 0, func(text string) {
			mit = text
		})
	form.SetTitle(pageAddTextRecord).
		SetTitleAlign(tview.AlignLeft)

	buttons := tview.NewForm().
		AddButton(buttonOkDesc, func() {
			textType := &models.Text{
				Data: text,
			}
			rdto, err := models.NewRecordDTO(
				desc,
				models.TextType,
				textType,
				metadata,
			)

			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			rows := splitMetadata(mit)
			m, err := models.NewMetadataFromStringArray(rows)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			rdto.Metadata = m

			_, err = ui.authUser.AddRecord(ctx, ui.cache, rdto)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			ui.pages.RemovePage(pageAddTextRecord)
		}).
		AddButton(buttonCancelDesc, func() { ui.pages.RemovePage(pageAddTextRecord) })

	buttons.SetButtonsAlign(tview.AlignLeft).SetBorderPadding(0, 0, 0, 0)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).AddItem(buttons, 1, 1, false)

	ui.pages.AddPage(pageAddTextRecord, flex, true, true)
}

func (ui *TUI) displayUpdateText(ctx context.Context, recordID string) {
	r, err := ui.authUser.GetRecord(ctx, ui.cache, recordID)
	if err != nil {
		ui.displayErr(err.Error())
	}

	t := &models.Text{}
	if err := cbor.Unmarshal(r.Data, t); err != nil {
		ui.displayErr(fmt.Sprintf("unmarshal data text binary, err: %v", err))
		return
	}

	desc := r.Description
	text := t.Data
	metadata := r.Metadata
	mit := convertMetadataToString(r.Metadata)

	form := tview.NewForm().
		AddInputField(fnDescription, r.Description, defaultFieldWidth, nil, func(v string) {
			desc = v
		}).
		AddInputField(fnText, t.Data, defaultFieldWidth, nil, func(v string) {
			text = v
		}).
		AddTextArea(fnMetadata, mit, defaultFieldWidth, 0, 0, func(text string) {
			mit = text
		})
	form.SetTitle(pageUpdateTextRecord).
		SetTitleAlign(tview.AlignLeft)

	buttons := tview.NewForm().
		AddButton(buttonUpdate, func() {
			textType := &models.Text{
				Data: text,
			}

			r, err := models.NewRecord(
				r.ID,
				desc,
				models.TextType,
				r.Created,
				time.Now(),
				textType,
				metadata,
				false,
				r.Version,
			)
			if err != nil {
				ui.displayErr(err.Error())
			}

			rows := splitMetadata(mit)
			m, err := models.NewMetadataFromStringArray(rows)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			r.Metadata = m

			r.Version++

			_, err = ui.authUser.UpdateRecord(ctx, ui.cache, r)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			ui.pages.RemovePage(pageUpdateTextRecord)
		}).
		AddButton(buttonCancelDesc, func() { ui.pages.RemovePage(pageUpdateTextRecord) })

	buttons.SetButtonsAlign(tview.AlignLeft).SetBorderPadding(0, 0, 0, 0)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).AddItem(buttons, 1, 1, false)

	ui.pages.AddPage(pageUpdateTextRecord, flex, true, true)
}

func (ui *TUI) displayCreateBinary(ctx context.Context) {
	var path string
	var desc string
	var metadata []*models.Metadata
	mit := ""
	form := tview.NewForm().
		AddInputField(fnDescription, "", defaultFieldWidth, nil, func(v string) {
			desc = v
		}).
		AddInputField(fnPath, "", defaultFieldWidth, nil, func(v string) {
			path = v
		}).
		AddTextArea(fnMetadata, mit, defaultFieldWidth, 0, 0, func(text string) {
			mit = text
		})
	form.SetTitle(pageAddBinaryRecord).
		SetTitleAlign(tview.AlignLeft)

	buttons := tview.NewForm().
		AddButton(buttonOkDesc, func() {
			f, err := os.ReadFile(path)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			s, err := os.Stat(path)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			if s.Size() > int64(models.MaxFileSize) {
				ui.displayErr(fmt.Sprintf("the file size should not exceed %d bytes", models.MaxFileSize))
				return
			}

			binaryType := &models.Binary{
				Data: f,
				Name: s.Name(),
				Ext:  filepath.Ext(path),
			}

			rdto, err := models.NewRecordDTO(
				desc,
				models.BinaryType,
				binaryType,
				metadata,
			)

			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			rows := splitMetadata(mit)
			m, err := models.NewMetadataFromStringArray(rows)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			rdto.Metadata = m

			_, err = ui.authUser.AddRecord(ctx, ui.cache, rdto)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			ui.pages.RemovePage(pageAddBinaryRecord)
		}).
		AddButton(buttonCancelDesc, func() { ui.pages.RemovePage(pageAddBinaryRecord) })

	buttons.SetButtonsAlign(tview.AlignLeft).SetBorderPadding(0, 0, 0, 0)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).AddItem(buttons, 1, 1, false)

	ui.pages.AddPage(pageAddBinaryRecord, flex, true, true)
}

func (ui *TUI) displayUpdateBinary(ctx context.Context, recordID string) {
	r, err := ui.authUser.GetRecord(ctx, ui.cache, recordID)
	if err != nil {
		ui.displayErr(err.Error())
	}

	bin := &models.Binary{}
	if err := cbor.Unmarshal(r.Data, bin); err != nil {
		ui.displayErr(fmt.Sprintf("unmarshal data binary, err: %v", err))
		return
	}

	desc := r.Description
	path := "Enter new path for file here..."
	metadata := r.Metadata
	mit := convertMetadataToString(r.Metadata)

	form := tview.NewForm().
		AddInputField(fnDescription, r.Description, defaultFieldWidth, nil, func(v string) {
			desc = v
		}).
		AddTextView("Filename", bin.Name, defaultFieldWidth, 1, false, true).
		AddTextView("Ext", bin.Ext, defaultFieldWidth, 1, false, true).
		AddButton("Save to OS", func() {
			if err := os.WriteFile(filepath.Join(path, bin.Name), bin.Data, defFileMode); err != nil {
				ui.displayErr(err.Error())
				return
			}
			ui.pages.RemovePage(pageUpdateBinaryRecord)
		}).
		AddInputField(fnPath, "", defaultFieldWidth, nil, func(v string) {
			path = v
		}).
		AddTextArea(fnMetadata, mit, defaultFieldWidth, 0, 0, func(text string) {
			mit = text
		})
	form.SetTitle(pageUpdateBinaryRecord).
		SetTitleAlign(tview.AlignLeft)

	buttons := tview.NewForm().
		AddButton(buttonUpdate, func() {
			f, err := os.ReadFile(path)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			s, err := os.Stat(path)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			if s.Size() > int64(models.MaxFileSize) {
				ui.displayErr(models.ErrLargeFile)
				return
			}

			binaryType := &models.Binary{
				Data: f,
				Name: s.Name(),
				Ext:  filepath.Ext(path),
			}

			r, err := models.NewRecord(
				r.ID,
				desc,
				models.BinaryType,
				r.Created,
				time.Now(),
				binaryType,
				metadata,
				false,
				r.Version,
			)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			rows := splitMetadata(mit)
			m, err := models.NewMetadataFromStringArray(rows)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			r.Metadata = m
			r.Version++

			_, err = ui.authUser.UpdateRecord(ctx, ui.cache, r)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			ui.pages.RemovePage(pageUpdateBinaryRecord)
		}).
		AddButton(buttonCancelDesc, func() { ui.pages.RemovePage(pageUpdateBinaryRecord) })

	buttons.SetButtonsAlign(tview.AlignLeft).SetBorderPadding(0, 0, 0, 0)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).AddItem(buttons, 1, 1, false)

	ui.pages.AddPage(pageUpdateBinaryRecord, flex, true, true)
}

const tempTerm = "01/06"

func (ui *TUI) displayCreateCard(ctx context.Context) {
	var desc string
	var number string
	var term time.Time
	var owner string
	var metadata []*models.Metadata
	mit := ""
	form := tview.NewForm().
		AddInputField(fnDescription, "", defaultFieldWidth, nil, func(v string) {
			desc = v
		}).
		AddInputField(fnNumber, "", defaultFieldWidth, nil, func(v string) {
			number = v
		}).
		AddInputField(fnOwner, "", defaultFieldWidth, nil, func(v string) {
			owner = v
		}).
		AddInputField(fnTerm, "", defaultFieldWidth, nil, func(v string) {
			if len(v) < len(tempTerm) {
				return
			}
			t, err := time.Parse(tempTerm, v)
			if err != nil {
				ui.displayErr(fmt.Sprintf("an error occured while parse term in create card form, err: %v", err))
			}
			term = t
		}).
		AddTextView(fnTemplateTermDesc, fnTemplateHintTermDesc, defaultFieldWidth, 0, true, true).
		AddTextArea(fnMetadata, mit, defaultFieldWidth, 0, 0, func(text string) {
			mit = text
		})

	form.SetTitle(pageAddCardRecord).
		SetTitleAlign(tview.AlignLeft)

	buttons := tview.NewForm().
		AddButton(buttonOkDesc, func() {
			cardType := &models.Card{
				Number: number,
				Term:   term,
				Owner:  owner,
			}
			rdto, err := models.NewRecordDTO(
				desc,
				models.CardType,
				cardType,
				metadata,
			)

			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			rows := splitMetadata(mit)
			m, err := models.NewMetadataFromStringArray(rows)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			rdto.Metadata = m

			_, err = ui.authUser.AddRecord(ctx, ui.cache, rdto)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			ui.pages.RemovePage(pageAddCardRecord)
		}).
		AddButton(buttonCancelDesc, func() { ui.pages.RemovePage(pageAddCardRecord) })

	buttons.SetButtonsAlign(tview.AlignLeft).SetBorderPadding(0, 0, 0, 0)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).AddItem(buttons, 1, 1, false)

	ui.pages.AddPage(pageAddCardRecord, flex, true, true)
}

func (ui *TUI) displayUpdateCard(ctx context.Context, recordID string) {
	r, err := ui.authUser.GetRecord(ctx, ui.cache, recordID)
	if err != nil {
		ui.displayErr(err.Error())
	}

	c := &models.Card{}
	if err := cbor.Unmarshal(r.Data, c); err != nil {
		ui.displayErr(fmt.Sprintf("unmarshal data card binary, err: %v", err))
		return
	}

	desc := r.Description
	number := c.Number
	term := c.Term
	owner := c.Owner
	metadata := r.Metadata
	mit := convertMetadataToString(r.Metadata)

	form := tview.NewForm().
		AddInputField(fnDescription, desc, defaultFieldWidth, nil, func(v string) {
			desc = v
		}).
		AddInputField(fnNumber, number, defaultFieldWidth, nil, func(v string) {
			number = v
		}).
		AddInputField(fnOwner, owner, defaultFieldWidth, nil, func(v string) {
			owner = v
		}).
		AddInputField(fnTerm, term.Format(tempTerm), defaultFieldWidth, nil, func(v string) {
			if len(v) < len(tempTerm) {
				return
			}
			t, err := time.Parse(tempTerm, v)
			if err != nil {
				ui.displayErr(fmt.Sprintf("an error occured while parse term in update card form, err: %v", err))
			}
			term = t
		}).
		AddTextView(fnTemplateTermDesc, fnTemplateHintTermDesc, defaultFieldWidth, 0, true, true).
		AddTextArea(fnMetadata, mit, defaultFieldWidth, 0, 0, func(text string) {
			mit = text
		})

	form.SetTitle(pageUpdateCardRecord).
		SetTitleAlign(tview.AlignLeft)

	buttons := tview.NewForm().
		AddButton(buttonUpdate, func() {
			cardType := &models.Card{
				Number: number,
				Term:   term,
				Owner:  owner,
			}

			r, err := models.NewRecord(
				r.ID,
				desc,
				models.CardType,
				r.Created,
				time.Now(),
				cardType,
				metadata,
				false,
				r.Version,
			)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			rows := splitMetadata(mit)
			m, err := models.NewMetadataFromStringArray(rows)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}
			r.Metadata = m

			r.Version++

			_, err = ui.authUser.UpdateRecord(ctx, ui.cache, r)
			if err != nil {
				ui.displayErr(err.Error())
				return
			}

			ui.pages.RemovePage(pageUpdateCardRecord)
		}).
		AddButton(buttonCancelDesc, func() { ui.pages.RemovePage(pageUpdateCardRecord) })

	buttons.SetButtonsAlign(tview.AlignLeft).SetBorderPadding(0, 0, 0, 0)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).AddItem(buttons, 1, 1, false)

	ui.pages.AddPage(pageUpdateCardRecord, flex, true, true)
}

func splitMetadata(md string) []string {
	return strings.Split(md, "\n")
}

func convertMetadataToString(md []*models.Metadata) string {
	mit := ""

	for _, m := range md {
		mit += fmt.Sprintf("%s:%s\n", m.Key, m.Value)
	}

	return mit
}
