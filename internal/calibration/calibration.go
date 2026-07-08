// Package calibration provides Brier/Murphy decomposition and ROC-AUC
// probabilistic forecast verification primitives for continuous estimates.
//
// These primitives are ported from the gaia ecosystem calibration suite
// (x-poll from gaia) and adapted for Go idioms.  They are used to wrap any
// floating-point estimate (e.g. the BNG biodiversity-unit gain) in a
// calibrated uncertainty envelope, telling the caller how far to trust the
// number given metric-version or measurement drift.
//
// Murphy (1973) decomposition:
//
//	BS = Reliability - Resolution + Uncertainty
//
// where BS is the mean Brier score, Reliability captures mean-squared
// deviation of forecast probabilities from observed frequencies (lower =
// better), Resolution captures how much conditional observed frequencies vary
// from the base rate (higher = more skill), and Uncertainty is an irreducible
// term equal to base_rate*(1-base_rate).
//
// ROC-AUC summarises the model's ability to rank positives above negatives
// across all possible thresholds:  0.5 = random, 1.0 = perfect.
//
// INFORMATIONAL ONLY — not a substitute for a licenced environmental assessor.
package calibration

import (
	"fmt"
	"math"
	"sort"
)

// ---------------------------------------------------------------------------
// Brier score
// ---------------------------------------------------------------------------

// BrierScore computes the mean Brier score (mean squared error of probability
// forecasts vs binary outcomes).
//
// probabilities: slice of forecast probabilities in [0, 1].
// observations:  slice of observed event indicators (any nonzero = 1).
//
// Lower is better; 0 = perfect; 0.25 = climatology at 50% base rate.
// Returns 0 for empty inputs.
func BrierScore(probabilities, observations []float64) (float64, error) {
	if len(probabilities) != len(observations) {
		return 0, fmt.Errorf("calibration: length mismatch: %d probs vs %d obs",
			len(probabilities), len(observations))
	}
	n := len(probabilities)
	if n == 0 {
		return 0.0, nil
	}
	total := 0.0
	for i := 0; i < n; i++ {
		p := clip01(probabilities[i])
		o := binarise(observations[i])
		d := p - o
		total += d * d
	}
	return total / float64(n), nil
}

// ---------------------------------------------------------------------------
// Murphy (1973) decomposition
// ---------------------------------------------------------------------------

// BrierDecomposition holds the Murphy (1973) three-way Brier-score
// decomposition and derived statistics.
type BrierDecomposition struct {
	// Reliability: mean squared bias of binned forecasts vs observed
	// frequencies.  0 = perfectly reliable.
	Reliability float64
	// Resolution: how much conditional observed frequencies vary from base rate.
	// Higher = more skill.
	Resolution float64
	// Uncertainty: irreducible term = base_rate * (1 - base_rate).
	Uncertainty float64
	// BaseRate: unconditional observed event frequency.
	BaseRate float64
	// NEvents: number of forecast/observation pairs.
	NEvents int
}

// BrierScore reconstructs the Brier score from the decomposition:
// BS = Reliability - Resolution + Uncertainty.
func (d BrierDecomposition) BrierScore() float64 {
	return d.Reliability - d.Resolution + d.Uncertainty
}

// Skill returns Resolution - Reliability; positive means the forecast beats
// climatology.
func (d BrierDecomposition) Skill() float64 {
	return d.Resolution - d.Reliability
}

// BrierDecompose performs the Murphy (1973) decomposition using n-bin equal-
// width probability bins.  nBins must be >= 1.
func BrierDecompose(probabilities, observations []float64, nBins int) (BrierDecomposition, error) {
	if len(probabilities) != len(observations) {
		return BrierDecomposition{}, fmt.Errorf("calibration: length mismatch: %d probs vs %d obs",
			len(probabilities), len(observations))
	}
	if nBins < 1 {
		return BrierDecomposition{}, fmt.Errorf("calibration: nBins must be >= 1, got %d", nBins)
	}
	n := len(probabilities)
	if n == 0 {
		return BrierDecomposition{}, nil
	}

	obsBin := make([]float64, n)
	baseRateSum := 0.0
	for i, o := range observations {
		obsBin[i] = binarise(o)
		baseRateSum += obsBin[i]
	}
	baseRate := baseRateSum / float64(n)
	uncertainty := baseRate * (1.0 - baseRate)

	binWidth := 1.0 / float64(nBins)
	binProbSum := make([]float64, nBins)
	binObsSum := make([]float64, nBins)
	binCount := make([]int, nBins)

	for i := 0; i < n; i++ {
		p := clip01(probabilities[i])
		idx := int(p / binWidth)
		if idx >= nBins {
			idx = nBins - 1
		}
		binProbSum[idx] += p
		binObsSum[idx] += obsBin[i]
		binCount[idx]++
	}

	reliability := 0.0
	resolution := 0.0
	for k := 0; k < nBins; k++ {
		if binCount[k] == 0 {
			continue
		}
		pAvg := binProbSum[k] / float64(binCount[k])
		oAvg := binObsSum[k] / float64(binCount[k])
		weight := float64(binCount[k]) / float64(n)
		reliability += weight * (pAvg - oAvg) * (pAvg - oAvg)
		resolution += weight * (oAvg - baseRate) * (oAvg - baseRate)
	}

	return BrierDecomposition{
		Reliability: reliability,
		Resolution:  resolution,
		Uncertainty: uncertainty,
		BaseRate:    baseRate,
		NEvents:     n,
	}, nil
}

// BrierSkillScore returns the Brier Skill Score against the climatology
// reference (constant forecast at the observed base rate).
//
//	BSS = 1 - BS / BS_ref
//
// Returns 1.0 for a perfect forecast, 0.0 for one tied with climatology,
// negative for one worse than climatology.  Returns 0.0 for empty inputs.
func BrierSkillScore(probabilities, observations []float64) (float64, error) {
	if len(probabilities) != len(observations) {
		return 0, fmt.Errorf("calibration: length mismatch: %d probs vs %d obs",
			len(probabilities), len(observations))
	}
	if len(probabilities) == 0 {
		return 0.0, nil
	}
	bs, err := BrierScore(probabilities, observations)
	if err != nil {
		return 0, err
	}
	// Climatology reference: constant forecast at base rate.
	obsSum := 0.0
	for _, o := range observations {
		obsSum += binarise(o)
	}
	baseRate := obsSum / float64(len(observations))
	ref := make([]float64, len(observations))
	for i := range ref {
		ref[i] = baseRate
	}
	bsRef, err := BrierScore(ref, observations)
	if err != nil {
		return 0, err
	}
	if bsRef < 1e-15 {
		if bs < 1e-15 {
			return 0.0, nil
		}
		return math.Inf(-1), nil
	}
	return 1.0 - bs/bsRef, nil
}

// ---------------------------------------------------------------------------
// ROC curve + AUC
// ---------------------------------------------------------------------------

// ROCCurve computes the Receiver Operating Characteristic curve by sweeping
// the probability threshold from high to low.
//
// Returns (fpr, tpr) — two slices of equal length, always with endpoints
// (0,0) and (1,1).  Ties in probability are grouped together.
// Empty inputs return the diagonal ([0,1],[0,1]).
func ROCCurve(probabilities, observations []float64) (fpr, tpr []float64, err error) {
	if len(probabilities) != len(observations) {
		return nil, nil, fmt.Errorf("calibration: length mismatch: %d probs vs %d obs",
			len(probabilities), len(observations))
	}
	n := len(probabilities)
	if n == 0 {
		return []float64{0.0, 1.0}, []float64{0.0, 1.0}, nil
	}

	type pair struct {
		p float64
		o int
	}
	pairs := make([]pair, n)
	posTotal := 0
	for i := 0; i < n; i++ {
		o := 0
		if observations[i] != 0 {
			o = 1
		}
		pairs[i] = pair{probabilities[i], o}
		posTotal += o
	}
	negTotal := n - posTotal

	// Sort descending by probability.
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].p > pairs[j].p })

	fprOut := []float64{0.0}
	tprOut := []float64{0.0}
	tp := 0
	fp := 0

	i := 0
	for i < n {
		sameProb := pairs[i].p
		j := i
		for j < n && pairs[j].p == sameProb {
			if pairs[j].o == 1 {
				tp++
			} else {
				fp++
			}
			j++
		}
		curTPR := 0.0
		if posTotal > 0 {
			curTPR = float64(tp) / float64(posTotal)
		}
		curFPR := 0.0
		if negTotal > 0 {
			curFPR = float64(fp) / float64(negTotal)
		}
		fprOut = append(fprOut, curFPR)
		tprOut = append(tprOut, curTPR)
		i = j
	}

	// Ensure endpoint (1, 1) for degenerate single-class cases.
	if fprOut[len(fprOut)-1] < 1.0 || tprOut[len(tprOut)-1] < 1.0 {
		fprOut = append(fprOut, 1.0)
		tprOut = append(tprOut, 1.0)
	}
	return fprOut, tprOut, nil
}

// ROCAUC computes the area under the ROC curve via the trapezoid rule.
//
// AUC = 0.5 means no discrimination; AUC = 1.0 means perfect.
// Empty inputs and degenerate single-class observations return 0.5.
func ROCAUC(probabilities, observations []float64) (float64, error) {
	if len(probabilities) != len(observations) {
		return 0, fmt.Errorf("calibration: length mismatch: %d probs vs %d obs",
			len(probabilities), len(observations))
	}
	if len(probabilities) == 0 {
		return 0.5, nil
	}
	posSum := 0.0
	for _, o := range observations {
		if o != 0 {
			posSum++
		}
	}
	if posSum == 0 || posSum == float64(len(observations)) {
		return 0.5, nil // Only one class — undefined; convention 0.5.
	}

	fprSlice, tprSlice, err := ROCCurve(probabilities, observations)
	if err != nil {
		return 0, err
	}
	auc := 0.0
	for k := 1; k < len(fprSlice); k++ {
		dx := fprSlice[k] - fprSlice[k-1]
		if dx <= 0.0 {
			continue
		}
		auc += 0.5 * dx * (tprSlice[k] + tprSlice[k-1])
	}
	return auc, nil
}

// ---------------------------------------------------------------------------
// GainUncertaintyBand — continuous-estimate uncertainty envelope for BNG gain
// ---------------------------------------------------------------------------

// GainUncertaintyBand wraps a continuous BNG percentage-gain estimate in an
// uncertainty envelope derived from the Brier/Murphy calibration of prior
// forecast/observation pairs.
//
// The "reliability" component of the Murphy decomposition is the mean squared
// deviation between forecast probabilities and observed frequencies; it is a
// direct analogue of mean-squared forecast error for the threshold-exceedance
// framing that the BNG gate uses (did the development meet the +10% threshold?).
// A low Reliability (< 0.05) means the forecasting model is well-calibrated;
// a high value signals metric-version drift or measurement uncertainty that
// the caller should factor in.
//
// halfWidth is the ±one-sigma uncertainty band on the stated gainPct, derived
// as sqrt(Reliability) scaled to the gain space.  When insufficient history
// is available halfWidth is 0 and CalibratedTrust is Low.
type GainUncertaintyBand struct {
	// GainPct is the raw BNG gain percentage (from Classify).
	GainPct float64
	// HalfWidth is the ±uncertainty (in percentage points).
	// Example: GainPct=12.5, HalfWidth=1.2 -> [11.3%, 13.7%].
	HalfWidth float64
	// CalibratedTrust reflects how much the caller should rely on GainPct.
	CalibratedTrust TrustLevel
	// BrierScore is the raw BS for reference (lower = better).
	BrierScore float64
	// SkillScore is the Brier Skill Score vs climatology (>0 = beats chance).
	SkillScore float64
	// ROCAUC is the ROC area under curve.
	ROCAUC float64
}

// TrustLevel conveys how well-calibrated the gain estimate is.
type TrustLevel string

const (
	// TrustHigh — reliability < 0.05; the gain is well-calibrated.
	TrustHigh TrustLevel = "HIGH"
	// TrustMedium — reliability in [0.05, 0.15).
	TrustMedium TrustLevel = "MEDIUM"
	// TrustLow — reliability >= 0.15 or insufficient history.
	TrustLow TrustLevel = "LOW"
)

// EnsembleExceedanceProb returns the empirical probability that any ensemble
// member exceeds the threshold (strict >).  Returns 0 for an empty ensemble.
// This matches the WMO Standard Verification System convention used by gaia.
func EnsembleExceedanceProb(members []float64, threshold float64) float64 {
	n := len(members)
	if n == 0 {
		return 0.0
	}
	count := 0
	for _, m := range members {
		if m > threshold {
			count++
		}
	}
	return float64(count) / float64(n)
}

// CalibrateBNGGain wraps a raw BNG gain percentage in an uncertainty envelope.
//
// priorProbabilities and priorObservations are the historical record of
// probabilistic threshold-exceedance forecasts and their binary outcomes
// (1 = threshold exceeded, 0 = not).  These should come from the site's or
// methodology's own calibration history.
//
// If fewer than minHistory pairs are available the band collapses to zero and
// trust is Low.  minHistory = 0 disables the guard.
func CalibrateBNGGain(
	gainPct float64,
	priorProbabilities []float64,
	priorObservations []float64,
	minHistory int,
) (GainUncertaintyBand, error) {
	band := GainUncertaintyBand{GainPct: gainPct}

	n := len(priorProbabilities)
	if minHistory > 0 && n < minHistory {
		band.CalibratedTrust = TrustLow
		return band, nil
	}
	if n == 0 {
		band.CalibratedTrust = TrustLow
		return band, nil
	}

	bs, err := BrierScore(priorProbabilities, priorObservations)
	if err != nil {
		return band, err
	}
	band.BrierScore = bs

	decomp, err := BrierDecompose(priorProbabilities, priorObservations, 10)
	if err != nil {
		return band, err
	}

	// Half-width: root-reliability in probability space; scale to gain-pct
	// space by the statutory minimum (10pp is our canonical unit).
	halfWidthProb := math.Sqrt(decomp.Reliability)
	const gainScalePP = 10.0 // 1 probability unit ≈ 10 percentage-point gain
	band.HalfWidth = halfWidthProb * gainScalePP

	// Trust classification.
	switch {
	case decomp.Reliability < 0.05:
		band.CalibratedTrust = TrustHigh
	case decomp.Reliability < 0.15:
		band.CalibratedTrust = TrustMedium
	default:
		band.CalibratedTrust = TrustLow
	}

	bss, err := BrierSkillScore(priorProbabilities, priorObservations)
	if err != nil {
		return band, err
	}
	if !math.IsInf(bss, 0) {
		band.SkillScore = bss
	}

	auc, err := ROCAUC(priorProbabilities, priorObservations)
	if err != nil {
		return band, err
	}
	band.ROCAUC = auc

	return band, nil
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func clip01(p float64) float64 {
	if p < 0.0 {
		return 0.0
	}
	if p > 1.0 {
		return 1.0
	}
	return p
}

func binarise(o float64) float64 {
	if o != 0 {
		return 1.0
	}
	return 0.0
}
