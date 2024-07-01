package strand

import "github.com/tinne26/ggfnt"

// The [StrandMapping] struct provides access to the rules
// for mapping code points to font glyphs. This involves
// rather advanced features for [*ggfnt.Font] objects, like
// rewrite rules, which I won't be explaining here right
// now (see ggfnt's spec directly if necessary).
//
// In general, this type is used through method chaining:
//   strand.Mapping().SetRewriteRulesEnabled(true)
type StrandMapping Strand

// Gateway to [StrandMapping].
func (self *Strand) Mapping() *StrandMapping {
	return (*StrandMapping)(self)
}

// Enables or disables rewrite rule processing for the strand.
// Even if disabled, previously configured rules remain stored
// and can be re-enabled at a later point.
func (self *StrandMapping) SetRewriteRulesEnabled(enabled bool) {
	if self.utf8Tester.IsOperating() || self.glyphTester.IsOperating() {
		panic("can't enable/disable rewrite rules while operating")
	}
	(*Strand)(self).setFlag(strandRewriteRulesDisabled, !enabled)
}

// Returns whether rewrite rules are enabled or not.
// See [StrandMapping.SetRewriteRulesEnabled]().
func (self *StrandMapping) GetRewriteRulesEnabled() bool {
	return !(*Strand)(self).getFlag(strandRewriteRulesDisabled)
}

// Performance note: when a rewrite rule is added, the decision tree in
// charge of evaluating glyph sequences needs to be recompiled. This
// will happen automatically the next time the strand is used, or it
// can be done preemptively through [StrandMapping.ResyncRewriteRules]().
//
// Notice: you must [StrandMapping.SetRewriteRulesEnabled](true) for
// added rules to take effect.
func (self *StrandMapping) AddUtf8RewriteRule(rule ggfnt.Utf8RewriteRule) error {
	return self.utf8Tester.AddRule(rule)
}

// Searches for the given rule (linearly) and removes it from the
// decision tree if found.
func (self *StrandMapping) DeleteUtf8RewriteRule(rule ggfnt.Utf8RewriteRule) bool {
	return self.utf8Tester.RemoveRule(rule)
}

// Performance note: when a rewrite rule is added, the decision tree in
// charge of evaluating glyph sequences needs to be recompiled. This
// will happen automatically the next time the strand is used, or it
// can be done preemptively through [StrandMapping.ResyncRewriteRules]().
//
// Notice: you must [StrandMapping.SetRewriteRulesEnabled](true) for
// added rules to take effect.
func (self *StrandMapping) AddGlyphRewriteRule(rule ggfnt.GlyphRewriteRule) error {
	return self.glyphTester.AddRule(rule)
}

// Searches for the given rule (linearly) and removes it from the
// decision tree if found.
func (self *StrandMapping) DeleteGlyphRewriteRule(rule ggfnt.GlyphRewriteRule) bool {
	return self.glyphTester.RemoveRule(rule)
}

// Manually requests a resync of the current rewrite rules. If the rules are
// already synced, nothing will happen; otherwise, the decision trees for the
// current rewrite rules will be recompiled.
// 
// Although manual resyncs can be used during preloads, notice that this
// is not a necessary function; most of the time, you can let the system
// recompile automatically as needed.
func (self *StrandMapping) ResyncRewriteRules() error {
	err := self.utf8Tester.Resync(self.font, &self.settings)
	if err != nil { return err }
	return self.glyphTester.Resync(self.font, &self.settings)
}

// Removes all existing rewrite rules.
func (self *StrandMapping) ClearAllRewriteRules() {
	self.glyphTester.RemoveAllRules()
	self.utf8Tester.RemoveAllRules()
}

// Utility method to automatically initialize and enable rewrite rules.
//
// Internally, this method looks up all font rules (both utf8 and glyph
// based) and adds them one by one to the rule testers. If any other rules
// were already present, they won't be removed.
func (self *StrandMapping) AutoInitRewriteRules() error {
	rewrites := self.font.Rewrites()
	for i := uint16(0); i < rewrites.NumGlyphRules(); i++ {
		err := self.AddGlyphRewriteRule(rewrites.GetGlyphRule(i))
		if err != nil { return err }
	}
	for i := uint16(0); i < rewrites.NumUTF8Rules(); i++ {
		err := self.AddUtf8RewriteRule(rewrites.GetUtf8Rule(i))
		if err != nil { return err }
	}
	return nil
}
