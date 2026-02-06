package state

import (
	"testing"
)

func TestQuickModeStages(t *testing.T) {
	if len(QuickStages) != 2 {
		t.Errorf("QuickStages should have 2 stages, got %d", len(QuickStages))
	}
	
	if QuickStages[0] != "requirements" {
		t.Errorf("First quick stage should be 'requirements', got %s", QuickStages[0])
	}
	
	if QuickStages[1] != "implementation" {
		t.Errorf("Second quick stage should be 'implementation', got %s", QuickStages[1])
	}
}

func TestNewQuickMode(t *testing.T) {
	task := "build a CLI tool"
	confidence := 70
	
	st := NewQuickMode(task, confidence)
	
	if !st.QuickMode {
		t.Error("QuickMode should be true")
	}
	
	if st.QuickTask != task {
		t.Errorf("QuickTask should be %q, got %q", task, st.QuickTask)
	}
	
	if st.Confidence != confidence {
		t.Errorf("Confidence should be %d, got %d", confidence, st.Confidence)
	}
	
	if st.CurrentStage != "requirements" {
		t.Errorf("CurrentStage should be 'requirements', got %s", st.CurrentStage)
	}
	
	// Should only have 2 gates
	if len(st.Gates) != 2 {
		t.Errorf("QuickMode should have 2 gates, got %d", len(st.Gates))
	}
	
	if st.Gates["requirements"] == nil {
		t.Error("Should have requirements gate")
	}
	
	if st.Gates["implementation"] == nil {
		t.Error("Should have implementation gate")
	}
	
	// Design and phases gates should not exist
	if st.Gates["design"] != nil {
		t.Error("Should not have design gate in quick mode")
	}
	
	if st.Gates["phases"] != nil {
		t.Error("Should not have phases gate in quick mode")
	}
}

func TestGetActiveStages(t *testing.T) {
	t.Run("full mode", func(t *testing.T) {
		st := NewDefault()
		stages := st.GetActiveStages()
		
		if len(stages) != 4 {
			t.Errorf("Full mode should have 4 stages, got %d", len(stages))
		}
	})
	
	t.Run("quick mode", func(t *testing.T) {
		st := NewQuickMode("test", 70)
		stages := st.GetActiveStages()
		
		if len(stages) != 2 {
			t.Errorf("Quick mode should have 2 stages, got %d", len(stages))
		}
	})
}

func TestGetNextStageForMode(t *testing.T) {
	t.Run("quick mode requirements to implementation", func(t *testing.T) {
		next := GetNextStageForMode("requirements", true)
		if next != "implementation" {
			t.Errorf("Expected 'implementation', got %q", next)
		}
	})
	
	t.Run("quick mode implementation is final", func(t *testing.T) {
		next := GetNextStageForMode("implementation", true)
		if next != "" {
			t.Errorf("Expected empty string (final stage), got %q", next)
		}
	})
	
	t.Run("full mode requirements to design", func(t *testing.T) {
		next := GetNextStageForMode("requirements", false)
		if next != "design" {
			t.Errorf("Expected 'design', got %q", next)
		}
	})
}

func TestQuickModeAdvance(t *testing.T) {
	st := NewQuickMode("test task", 70)
	
	// Approve requirements
	err := st.ApproveGate("requirements", "auto")
	if err != nil {
		t.Fatalf("Failed to approve requirements: %v", err)
	}
	
	// Should now be at implementation
	if st.CurrentStage != "implementation" {
		t.Errorf("Should be at 'implementation', got %s", st.CurrentStage)
	}
	
	// Requirements should be approved
	if st.Gates["requirements"].Status != "approved" {
		t.Error("Requirements gate should be approved")
	}
	
	// Implementation should be open
	if st.Gates["implementation"].Status != "open" {
		t.Errorf("Implementation gate should be 'open', got %s", st.Gates["implementation"].Status)
	}
}

func TestShouldAutoAdvance(t *testing.T) {
	t.Run("no auto-advance configured", func(t *testing.T) {
		st := NewDefault()
		if st.ShouldAutoAdvance(100) {
			t.Error("Should not auto-advance when confidence is 0")
		}
	})
	
	t.Run("below threshold", func(t *testing.T) {
		st := NewQuickMode("test", 80)
		if st.ShouldAutoAdvance(70) {
			t.Error("Should not auto-advance below threshold")
		}
	})
	
	t.Run("at threshold", func(t *testing.T) {
		st := NewQuickMode("test", 80)
		if !st.ShouldAutoAdvance(80) {
			t.Error("Should auto-advance at threshold")
		}
	})
	
	t.Run("above threshold", func(t *testing.T) {
		st := NewQuickMode("test", 80)
		if !st.ShouldAutoAdvance(90) {
			t.Error("Should auto-advance above threshold")
		}
	})
}
