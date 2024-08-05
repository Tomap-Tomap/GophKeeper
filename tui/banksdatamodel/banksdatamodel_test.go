//go:build unit

package banksdatamodel

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

const (
	testID     = "test ID"
	testName   = "John Doe"
	testNumber = "1234 5678 9012 3456"
	testCvc    = "123"
	testOwner  = "JOHN DOE"
	testExp    = "12/34"
	testMeta   = "test Meta"
)

type testEnterMsg struct {
}

func testEnterCmd() tea.Msg {
	return testEnterMsg{}
}

func TestNew(t *testing.T) {
	model := New(testEnterCmd, testID, testName, testNumber, testCvc, testOwner, testExp, testMeta)

	assert.Equal(t, testID, model.id)
	assert.Equal(t, testName, model.inputs[0].Value())
	assert.Equal(t, testNumber, model.inputs[1].Value())
	assert.Equal(t, testCvc, model.inputs[2].Value())
	assert.Equal(t, testOwner, model.inputs[3].Value())
	assert.Equal(t, testExp, model.inputs[4].Value())
	assert.Equal(t, testMeta, model.inputs[5].Value())
}

func TestInput(t *testing.T) {
	model := New(testEnterCmd, testID, testName, testNumber, testCvc, testOwner, testExp, testMeta)

	cmd := model.Init()

	assert.Nil(t, cmd)
}

func TestUpdate(t *testing.T) {
	model := New(testEnterCmd, testID, testName, testNumber, testCvc, testOwner, testExp, testMeta)

	t.Run("key down test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ := model.Update(msg)
		assert.Equal(t, 1, updatedModel.focused)
	})

	t.Run("key up test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyUp}
		updatedModel, _ := model.Update(msg)
		assert.Equal(t, 0, updatedModel.focused)
	})

	t.Run("enter test", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := model.Update(msg)
		assert.Equal(t, testEnterMsg{}, cmd())
	})

	t.Run("windows size test", func(t *testing.T) {
		msg := tea.WindowSizeMsg{
			Width:  100,
			Height: 100,
		}

		wantWidth := msg.Width - model.elementStyle.GetHorizontalFrameSize()
		wantHeight := msg.Height/numInputs - model.elementStyle.GetVerticalFrameSize()

		m, _ := model.Update(msg)
		assert.Equal(t, wantHeight, m.elementStyle.GetHeight())
		assert.Equal(t, wantHeight, m.focusedElementStyle.GetHeight())
		assert.Equal(t, wantWidth, m.elementStyle.GetWidth())
		assert.Equal(t, wantWidth, m.elementStyle.GetWidth())
	})
}

func TestView(t *testing.T) {
	model := New(testEnterCmd, testID, testName, testNumber, testCvc, testOwner, testExp, testMeta)

	view := model.View()
	assert.Contains(t, view, fmt.Sprintf("Name > %s", testName))
	assert.Contains(t, view, fmt.Sprintf("Card number > %s", testNumber))
	assert.Contains(t, view, fmt.Sprintf("CVC > %s", testCvc))
	assert.Contains(t, view, fmt.Sprintf("OWNER > %s", testOwner))
	assert.Contains(t, view, fmt.Sprintf("MM/YY > %s", testExp))
	assert.Contains(t, view, fmt.Sprintf("Meta > %s", testMeta))
}

func TestGetResult(t *testing.T) {
	model := New(testEnterCmd, testID, testName, testNumber, testCvc, testOwner, testExp, testMeta)

	resultID, resultName, resultNumber, resultCVC, resultOwner, resultExp, resultMeta := model.GetResult()

	assert.Equal(t, testID, resultID)
	assert.Equal(t, testName, resultName)
	assert.Equal(t, testNumber, resultNumber)
	assert.Equal(t, testCvc, resultCVC)
	assert.Equal(t, testOwner, resultOwner)
	assert.Equal(t, testExp, resultExp)
	assert.Equal(t, testMeta, resultMeta)
}

func TestCcnValidator(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid credit card number",
			input:   testNumber,
			wantErr: false,
		},
		{
			name:    "CCN is too long",
			input:   "4111 1111 1111 11111",
			wantErr: true,
		},
		{
			name:    "CCN is invalid",
			input:   "4111 1111 1111 111a",
			wantErr: true,
		},
		{
			name:    "CCN must separate groups with spaces",
			input:   "41111",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ccnValidator(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_cvvValidator(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "positive test",
			input:   testCvc,
			wantErr: false,
		},
		{
			name:    "error test",
			input:   "abc",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cvvValidator(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_expValidator(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "positive test",
			input:   testExp,
			wantErr: false,
		},
		{
			name:    "EXP is invalid",
			input:   "ab/bc",
			wantErr: true,
		},
		{
			name:    "EXP doesn't contain /",
			input:   "0123",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := expValidator(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_ownerValidator(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "positive test",
			input:   testOwner,
			wantErr: false,
		},
		{
			name:    "owner is invalid",
			input:   "s",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ownerValidator(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
