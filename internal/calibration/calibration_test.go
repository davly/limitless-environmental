package calibration

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// BrierScore
// ---------------------------------------------------------------------------

func TestBrierScore_Perfect(t *testing.T) {
	probs := []float64{1.0, 0.0, 1.0, 0.0}
	obs := []float64{1, 0, 1, 0}
	bs, err := BrierScore(probs, obs)
	if err != nil {
		t.Fatalf("BrierScore: %v", err)
	}
	if math.Abs(bs) > 1e-15 {
		t.Fatalf("perfect forecast: want 0, got %v", bs)
	}
}

func TestBrierScore_Worst(t *testing.T) {
	probs := []float64{0.0, 1.0, 0.0, 1.0}
	obs := []float64{1, 0, 1, 0}
	bs, err := BrierScore(probs, obs)
	if err != nil {
		t.Fatalf("BrierScore: %v", err)
	}
	if math.Abs(bs-1.0) > 1e-15 {
		t.Fatalf("worst forecast: want 1.0, got %v", bs)
	}
}

func TestBrierScore_KnownValue(t *testing.T) {
	// squares: 0.04 + 0.16 + 0.01 = 0.21; mean = 0.07
	probs := []float64{0.2, 0.6, 0.9}
	obs := []float64{0, 1, 1}
	bs, err := BrierScore(probs, obs)
	if err != nil {
		t.Fatalf("BrierScore: %v", err)
	}
	if math.Abs(bs-0.07) > 1e-12 {
		t.Fatalf("known value: want 0.07, got %v", bs)
	}
}

func TestBrierScore_Empty(t *testing.T) {
	bs, err := BrierScore(nil, nil)
	if err != nil {
		t.Fatalf("BrierScore empty: %v", err)
	}
	if bs != 0.0 {
		t.Fatalf("empty: want 0, got %v", bs)
	}
}

func TestBrierScore_LengthMismatch(t *testing.T) {
	_, err := BrierScore([]float64{0.5, 0.5}, []float64{1.0})
	if err == nil {
		t.Fatal("expected error for length mismatch, got nil")
	}
}

// ---------------------------------------------------------------------------
// BrierDecompose
// ---------------------------------------------------------------------------

func TestBrierDecompose_UncertaintyFormula(t *testing.T) {
	// base_rate 0.4 -> uncertainty = 0.4 * 0.6 = 0.24
	probs := make([]float64, 10)
	for i := range probs {
		probs[i] = 0.5
	}
	obs := []float64{1, 1, 1, 1, 0, 0, 0, 0, 0, 0}
	d, err := BrierDecompose(probs, obs, 10)
	if err != nil {
		t.Fatalf("BrierDecompose: %v", err)
	}
	if math.Abs(d.Uncertainty-0.24) > 1e-12 {
		t.Fatalf("uncertainty: want 0.24, got %v", d.Uncertainty)
	}
}

func TestBrierDecompose_ReconstructionApprox(t *testing.T) {
	// BS = Rel - Res + Unc; decomposition is binned so approximately equal.
	probs := []float64{0.9, 0.8, 0.7, 0.3, 0.2, 0.1}
	obs := []float64{1, 1, 1, 0, 0, 0}
	d, err := BrierDecompose(probs, obs, 10)
	if err != nil {
		t.Fatalf("BrierDecompose: %v", err)
	}
	bs, _ := BrierScore(probs, obs)
	if math.Abs(d.BrierScore()-bs) > 0.05 {
		t.Fatalf("reconstruction: BS=%v decomp=%v diff too large", bs, d.BrierScore())
	}
}

func TestBrierDecompose_Empty(t *testing.T) {
	d, err := BrierDecompose(nil, nil, 10)
	if err != nil {
		t.Fatalf("BrierDecompose empty: %v", err)
	}
	if d.Reliability != 0 || d.Resolution != 0 || d.Uncertainty != 0 {
		t.Fatalf("empty decomp should be all-zero, got %+v", d)
	}
}

func TestBrierDecompose_ZeroBinsError(t *testing.T) {
	_, err := BrierDecompose([]float64{0.5}, []float64{1.0}, 0)
	if err == nil {
		t.Fatal("expected error for nBins=0")
	}
}

// ---------------------------------------------------------------------------
// BrierSkillScore
// ---------------------------------------------------------------------------

func TestBrierSkillScore_Perfect(t *testing.T) {
	probs := []float64{1.0, 0.0, 1.0, 0.0}
	obs := []float64{1, 0, 1, 0}
	bss, err := BrierSkillScore(probs, obs)
	if err != nil {
		t.Fatalf("BrierSkillScore: %v", err)
	}
	if math.Abs(bss-1.0) > 1e-12 {
		t.Fatalf("perfect forecast: want 1.0, got %v", bss)
	}
}

func TestBrierSkillScore_Climatology(t *testing.T) {
	probs := []float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5}
	obs := []float64{1, 0, 1, 0, 1, 0, 1, 0, 1, 0}
	bss, err := BrierSkillScore(probs, obs)
	if err != nil {
		t.Fatalf("BrierSkillScore: %v", err)
	}
	if math.Abs(bss) > 1e-12 {
		t.Fatalf("climatology forecast: want 0, got %v", bss)
	}
}

func TestBrierSkillScore_Empty(t *testing.T) {
	bss, err := BrierSkillScore(nil, nil)
	if err != nil {
		t.Fatalf("BrierSkillScore empty: %v", err)
	}
	if bss != 0.0 {
		t.Fatalf("empty: want 0, got %v", bss)
	}
}

// ---------------------------------------------------------------------------
// ROCCurve + ROCAUC
// ---------------------------------------------------------------------------

func TestROCAUC_PerfectSeparation(t *testing.T) {
	probs := []float64{0.9, 0.8, 0.2, 0.1}
	obs := []float64{1, 1, 0, 0}
	auc, err := ROCAUC(probs, obs)
	if err != nil {
		t.Fatalf("ROCAUC: %v", err)
	}
	if math.Abs(auc-1.0) > 1e-12 {
		t.Fatalf("perfect separation: want 1.0, got %v", auc)
	}
}

func TestROCAUC_RandomAUCHalf(t *testing.T) {
	probs := []float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5}
	obs := []float64{1, 0, 1, 0, 1, 0}
	auc, err := ROCAUC(probs, obs)
	if err != nil {
		t.Fatalf("ROCAUC: %v", err)
	}
	if math.Abs(auc-0.5) > 1e-12 {
		t.Fatalf("random classifier: want 0.5, got %v", auc)
	}
}

func TestROCAUC_Empty(t *testing.T) {
	auc, err := ROCAUC(nil, nil)
	if err != nil {
		t.Fatalf("ROCAUC empty: %v", err)
	}
	if auc != 0.5 {
		t.Fatalf("empty: want 0.5, got %v", auc)
	}
}

func TestROCCurve_Endpoints(t *testing.T) {
	probs := []float64{0.3, 0.7, 0.5, 0.9}
	obs := []float64{0, 1, 1, 0}
	fpr, tpr, err := ROCCurve(probs, obs)
	if err != nil {
		t.Fatalf("ROCCurve: %v", err)
	}
	if fpr[0] != 0.0 || tpr[0] != 0.0 {
		t.Fatalf("ROCCurve must start at (0,0), got (%v,%v)", fpr[0], tpr[0])
	}
	last := len(fpr) - 1
	if fpr[last] != 1.0 || tpr[last] != 1.0 {
		t.Fatalf("ROCCurve must end at (1,1), got (%v,%v)", fpr[last], tpr[last])
	}
}

// ---------------------------------------------------------------------------
// EnsembleExceedanceProb
// ---------------------------------------------------------------------------

func TestEnsembleExceedanceProb_AllExceed(t *testing.T) {
	if EnsembleExceedanceProb([]float64{10.0, 11.0, 12.0}, 5.0) != 1.0 {
		t.Fatal("all members exceed: want 1.0")
	}
}

func TestEnsembleExceedanceProb_NoneExceed(t *testing.T) {
	if EnsembleExceedanceProb([]float64{1.0, 2.0, 3.0}, 100.0) != 0.0 {
		t.Fatal("no members exceed: want 0.0")
	}
}

func TestEnsembleExceedanceProb_StrictInequality(t *testing.T) {
	// Equal-to-threshold members do NOT count (strict >).
	p := EnsembleExceedanceProb([]float64{5.0, 5.0, 5.0, 6.0}, 5.0)
	if math.Abs(p-0.25) > 1e-12 {
		t.Fatalf("strict inequality at threshold: want 0.25, got %v", p)
	}
}

func TestEnsembleExceedanceProb_Empty(t *testing.T) {
	if EnsembleExceedanceProb(nil, 0.0) != 0.0 {
		t.Fatal("empty ensemble: want 0.0")
	}
}

// ---------------------------------------------------------------------------
// CalibrateBNGGain — the wired integration point
// ---------------------------------------------------------------------------

func TestCalibrateBNGGain_WellCalibrated(t *testing.T) {
	// Perfect calibration: always forecast 1.0, always observe 1. BS=0.
	probs := make([]float64, 40)
	obs := make([]float64, 40)
	for i := range probs {
		probs[i] = 1.0
		obs[i] = 1.0
	}
	band, err := CalibrateBNGGain(12.5, probs, obs, 10)
	if err != nil {
		t.Fatalf("CalibrateBNGGain: %v", err)
	}
	if band.GainPct != 12.5 {
		t.Fatalf("GainPct passthrough: want 12.5, got %v", band.GainPct)
	}
	if band.CalibratedTrust != TrustHigh {
		t.Fatalf("well-calibrated: want TrustHigh, got %v", band.CalibratedTrust)
	}
}

func TestCalibrateBNGGain_InsufficientHistory(t *testing.T) {
	band, err := CalibrateBNGGain(12.5, []float64{0.5}, []float64{1.0}, 10)
	if err != nil {
		t.Fatalf("CalibrateBNGGain insufficient: %v", err)
	}
	if band.CalibratedTrust != TrustLow {
		t.Fatalf("insufficient history: want TrustLow, got %v", band.CalibratedTrust)
	}
	if band.HalfWidth != 0.0 {
		t.Fatalf("insufficient history: want HalfWidth=0, got %v", band.HalfWidth)
	}
}

func TestCalibrateBNGGain_Miscalibrated(t *testing.T) {
	// Forecast 0.9 but event never happens — high Brier score, low trust.
	probs := make([]float64, 40)
	obs := make([]float64, 40)
	for i := range probs {
		probs[i] = 0.9
		obs[i] = 0.0
	}
	band, err := CalibrateBNGGain(10.5, probs, obs, 10)
	if err != nil {
		t.Fatalf("CalibrateBNGGain: %v", err)
	}
	// Brier score = 0.81; reliability will be high -> TrustLow
	if band.CalibratedTrust == TrustHigh {
		t.Fatalf("miscalibrated: should not be TrustHigh, got %v (BS=%v)", band.CalibratedTrust, band.BrierScore)
	}
	if band.HalfWidth <= 0.0 {
		t.Fatalf("miscalibrated: HalfWidth should be positive, got %v", band.HalfWidth)
	}
}

func TestCalibrateBNGGain_ZeroHistory(t *testing.T) {
	band, err := CalibrateBNGGain(10.0, nil, nil, 0)
	if err != nil {
		t.Fatalf("CalibrateBNGGain zero history: %v", err)
	}
	if band.CalibratedTrust != TrustLow {
		t.Fatalf("zero history: want TrustLow, got %v", band.CalibratedTrust)
	}
}
