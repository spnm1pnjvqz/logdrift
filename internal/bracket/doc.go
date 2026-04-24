// Package bracket provides a pipeline stage that wraps every log-line's text
// with a configurable open and close delimiter pair.
//
// Example usage:
//
//	b, err := bracket.New("[", "]")
//	if err != nil {
//		log.Fatal(err)
//	}
//	out := b.Apply(ctx, src)
package bracket
