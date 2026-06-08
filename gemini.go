package main

import (
	"context"
	"time"

	"google.golang.org/genai"
)

type geminiConnect struct {
	client *genai.Client
}

func GenaiNew() *geminiConnect {
	g := &geminiConnect{client: nil}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		messenger.AddLog(err.Error())
		return nil
	}
	g.client = client
	return g
}

func (g *geminiConnect) ask(question string) {
	messenger.Alert("info", "Asking Gemini, task will be sent to background ...")
	go func() {
		model := GetGlobalOption("geminimodel").(string)
		ctx := context.Background()
		thinkingBudgetVal := int32(0)
		var err error
		var result *genai.GenerateContentResponse
		for i := range 5 {
			result, err = g.client.Models.GenerateContent(
				ctx,
				model,
				genai.Text(question),
				&genai.GenerateContentConfig{
					ThinkingConfig: &genai.ThinkingConfig{
						ThinkingBudget: &thinkingBudgetVal,
					},
				})
			if err != nil && i == 5 {
				messenger.Alert("warning", "Gemini error, check log")
				messenger.AddLog(err)
				gemini = nil
				return
			} else if err != nil {
				messenger.AddLog("retry:", i, "-", err)
				time.Sleep(3 * time.Second)
			}
		}
		CurView().AddTab(false)
		CurView().Buf = NewBufferFromString(result.Text(), "")
		CurView().Buf.Settings["filetype"] = "gemini"
		CurView().Type = vtLog
		CurView().Buf.UpdateRules()
		CurView().Buf.Fname = "gemini"
		SetLocalOption("ruler", "false", CurView())
		SetLocalOption("softwrap", "true", CurView())
		navigationMode = true
		messenger.ClearMessage()
		RedrawAll(false)
	}()
}
