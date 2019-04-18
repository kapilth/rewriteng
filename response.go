package rewriteng

import (
	"net"
	"strings"

	"github.com/miekg/dns"
)

// ResponseRewriter rewrites answers, additional and authority sections
type ResponseRewriter struct {
	dns.ResponseWriter
	originalQuestion dns.Question
	Rules            []*nameRule
}

// NewResponseRewriter returns a pointer to a new ResponseRewriter.
func NewResponseRewriter(w dns.ResponseWriter, r *dns.Msg) *ResponseRewriter {
	return &ResponseRewriter{
		ResponseWriter:   w,
		originalQuestion: r.Question[0],
	}
}

// WriteMsg records the status code and calls the underlying ResponseWriter's WriteMsg method.
func (r *ResponseRewriter) WriteMsg(res *dns.Msg) error {
	res.Question[0] = r.originalQuestion
	// Answers
	r.rewriteAnswers(res)
	// Authority
	r.rewriteAuthority(res)
	// Additional
	r.rewriteAdditional(res)

	return r.ResponseWriter.WriteMsg(res)
}

func (r *ResponseRewriter) rewriteAnswers(res *dns.Msg) {
	for _, rr := range res.Answer {
		var nameWriten bool
		var name = rr.Header().Name
		for _, rule := range r.Rules {
			if !nameWriten {
				if s := rule.Sub(name, answerRule, namePart); s != "" {
					rr.Header().Name = s
					nameWriten = true
				}
			}
			// Rewrite Data
			rewriteDataParts(rr, rule, answerRule)
		}
	}
}

func (r *ResponseRewriter) rewriteAuthority(res *dns.Msg) {
	for _, rr := range res.Ns {
		var nameWriten bool
		var name = rr.Header().Name
		for _, rule := range r.Rules {
			if !nameWriten {
				if s := rule.Sub(name, authorityRule, namePart); s != "" {
					rr.Header().Name = s
					nameWriten = true
				}
			}
			// Rewrite Data
			rewriteDataParts(rr, rule, authorityRule)
		}
	}
}

func (r *ResponseRewriter) rewriteAdditional(res *dns.Msg) {
	for _, rr := range res.Extra {
		var nameWriten bool
		var name = rr.Header().Name
		for _, rule := range r.Rules {
			if !nameWriten {
				if s := rule.Sub(name, additionalRule, namePart); s != "" {
					rr.Header().Name = s
					nameWriten = true
				}
			}
			// Rewrite Data
			rewriteDataParts(rr, rule, additionalRule)
		}
	}
}

func rewriteDataParts(rr dns.RR, rule *nameRule, ruleType string) {
	switch t := rr.(type) {
	case *dns.CNAME:
		if s := rule.Sub(t.Target, ruleType, dataPart); s != "" {
			t.Target = s
		}
	case *dns.TXT:
		tmpRecs := t.Txt[:0]
		for _, txtr := range t.Txt {
			if s := rule.Sub(txtr, ruleType, dataPart); s != "" {
				tmpRecs = append(tmpRecs, s)
			} else {
				tmpRecs = append(tmpRecs, txtr)
			}
		}
		t.Txt = tmpRecs
	case *dns.A:
		if s := rule.Sub(t.A.String(), ruleType, dataPart); s != "" {
			if ip := net.ParseIP(s); ip != nil && strings.Contains(s, ".") {
				t.A = ip
			}
		}
	case *dns.AAAA:
		if s := rule.Sub(t.AAAA.String(), ruleType, dataPart); s != "" {
			if ip := net.ParseIP(s); ip != nil && strings.Contains(s, ":") {
				t.AAAA = ip
			}
		}
	case *dns.NS:
		if s := rule.Sub(t.Ns, ruleType, dataPart); s != "" {
			t.Ns = s
		}
	case *dns.SOA:
		if s := rule.Sub(t.Ns, ruleType, dataPart); s != "" {
			t.Ns = s
		}
		if s := rule.Sub(t.Mbox, ruleType, dataPart); s != "" {
			t.Mbox = s
		}
	case *dns.SRV:
		if s := rule.Sub(t.Target, ruleType, dataPart); s != "" {
			t.Target = s
		}
	}
}
