package framekit

import "math"

// Ease maps linear progress [0,1] to a premium-feeling curve.
func Ease(kind string, x float64) float64 {
	x = clamp01(x)
	switch kind {
	case "linear":
		return x
	case "easeIn":
		return x * x
	case "easeInOut":
		return -(math.Cos(math.Pi*x) - 1) / 2
	case "easeOutQuint":
		return 1 - math.Pow(1-x, 5)
	case "spring":
		if x == 1 {
			return 1
		}
		return 1 - math.Exp(-6*x)*math.Cos(10*x)
	default: // easeOutCubic
		return 1 - math.Pow(1-x, 3)
	}
}

// EasedReveal applies the configured easing to an appearance window.
func EasedReveal(spec Spec, t, at float64) float64 {
	if spec.Effects.Fade <= 0 {
		if t >= at {
			return 1
		}
		return 0
	}
	return Ease(spec.Effects.Easing, (t-at)/spec.Effects.Fade)
}

// EasedExit handles the end-of-scene fade.
func EasedExit(spec Spec, local, duration float64) float64 {
	if spec.Effects.Fade <= 0 || spec.Effects.Transition == "none" {
		return 1
	}
	return Ease(spec.Effects.Easing, (duration-local)/spec.Effects.Fade)
}
