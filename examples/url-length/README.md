# Worked example — "the maximum URL length is 2048"

This is the example the rest of the repo was missing. The others exercise the machinery; this one
shows the intended user — someone trained in close reading and dialectic — watching that training
help decisively on two of assay's three columns and turn into a trap on the third. The point is not
the verdict on the claim. The point is seeing where your own rigour stops.

## The claim

Something every developer has heard, stated the way it actually circulates (see `claim.txt`):

> The maximum length of a URL is 2048 characters, a hard limit defined by the HTTP
> specification; any longer URL is invalid.

and its popular correction, which over-shoots in the other direction:

> The HTTP specification places no limit, so a URL can be any length you want.

## The dispute (public substrate)

Trace the "2048" the whole crowd repeats and it descends from a single ancestor: Microsoft's
WinINET header, where `INTERNET_MAX_URL_LENGTH` is 2083 — Internet Explorer's limit, not the
web's. A 2010 measurement on the W3C `uri@w3.org` list found the sharp drop at ~2048 in
real-world URLs was caused by IE specifically; Firefox and Chrome showed no limit the author
could find up to 32k. The HTTP specification itself sets no hard upper bound on URI length and
only *recommends* supporting at least 8000 octets (RFC 2616 historically; RFC 9110/9112 today).

But "no limit in the spec" is not "no limit in practice": Apache's `LimitRequestLine` defaults to
~8190, proxies and CDNs cap header size, the Sitemaps protocol enforces 2048, and HTTP 414 (URI
Too Long) exists for exactly this. The real ceiling is a property of deployed systems, not of the
protocol.

Sources:
- Microsoft, "URL Length Limits" (IEInternals) — https://learn.microsoft.com/en-us/archive/blogs/ieinternals/url-length-limits
- W3C `uri@w3.org`, April 2010 — https://lists.w3.org/Archives/Public/uri/2010Apr/0005.html
- HTTP semantics, RFC 9110 — https://www.rfc-editor.org/rfc/rfc9110
- Apache `LimitRequestLine`; HTTP 414 URI Too Long; Sitemaps protocol (2048-char URL limit).

## Run it

```
./assay examples/url-length/claim.txt              # substance
./assay examples/url-length/claim.txt -evidence    # grounding (uses web_search)
```

The verdict tables are deliberately not pre-filled — see the note at the bottom for why.

## The three columns

**Faithfulness — helps.** "*The* maximum URL length is 2048" attributes a universal property to
URLs, and "defined by the HTTP specification" attributes it to a source that does not say it.
Close reading catches both: a vendor cap restated as a rule of the protocol, and a citation trail
that loops back to one IE constant. A thousand sources agreeing is not corroboration when they
share one origin — the trained reader sees the shared ancestor the fluent crowd misses.

**Substance — helps, and seduces.** The hidden premise is *implementation = specification*.
Dialectic pulls it out cleanly, and the moment it is named — "it's just IE, the spec sets no
limit" — it feels like the answer. That feeling is the trap. It is the armchair part, and it
feels finished.

**Grounding — the training does nothing, and over-confidence does harm.** Whoever stops at "the
spec sets no limit" and concludes "so use any length" has just asserted the *second* claim above,
and it is also false: real systems cap URLs (Apache ~8190, proxies, CDNs, sitemaps 2048, HTTP
414). The limit you must actually build against is a fact about deployed software, recoverable
only by going to the current RFC and the configs of the systems in your path — never by
reasoning, however rigorous. Both the myth and its clever debunking are armchair claims about the
world that the armchair cannot settle.

## The lesson

The same training wins decisively twice and is a trap the third time. The skill is not "apply
more rigour" — it is knowing the line runs between the second column and the third, and going to
the truth-maker once you are past it.

One extra tooth: even the debunking ("it's just IE, not the spec") only holds because someone
*read the spec*. That read was a retrieval, not a deduction. So the armchair feels finished one
move early, then feels finished again one move early. Knowing where your rigour stops includes
noticing the lookups you made without calling them lookups.

## A note on this file's own construction

The faithfulness and substance analyses above are authored from the armchair, because those
columns are reachable by reasoning. The grounding verdicts are not pre-filled, and that is
deliberate: the grounded answer is a retrieval, and writing plausible verdicts here without
running the tool would fabricate the one column the tool exists to protect. Run the commands and
paste your output below. The example refuses to invent its own grounding column — which is the
whole point.

<!-- VERDICT TABLES: paste real ./assay output here. Do not fabricate. -->

### Substance (`./assay examples/url-length/claim.txt -md`)

Run: 2026-06-01, model `claude-sonnet-4-6`, 7 claims, 15 calls, est $0.1334.

| Verdict | Claim                                                     | Note |
|---------|-----------------------------------------------------------|------|
| hollow  | The maximum length of a URL is 2048 characters            | Survives only by redefining 2048 as a "compatibility guideline for IE-era infrastructure" — conditions the speaker never stated. |
| hollow  | The 2048-character URL length limit is defined by the HTTP specification | Flatly false. No HTTP RFC defines a 2048-char limit. Steelman replaces "HTTP specification" with "broadly construed ecosystem including vendor documentation." Condition laundering. |
| hollow  | The HTTP specification defines a hard limit on URL length | RFC 7230 explicitly declines to set a numeric URI ceiling. Steelman redefines "hard limit" to mean "per-server discretionary threshold" — the opposite. |
| hollow  | Any URL longer than 2048 characters is invalid            | Universal claim trivially falsified by modern browsers and RFC 7230. Survives only as "in IE-era stacks explicitly configured to that limit." |
| hollow  | Any URL longer than 2048 characters will be rejected      | Same. Universal falsified by counterexample; rescued only with four conditions wholly absent from the original assertion. |
| partial | The HTTP specification places no limit on URL length      | **Survives as:** The IETF HTTP RFCs (RFC 7230 / RFC 9110–9112) specify no normative numeric maximum for URL length; any length ceiling is an implementation or deployment decision, not a protocol mandate. |
| partial | A URL can be any length                                   | **Survives as:** RFC 3986 and related URI specifications impose no explicit maximum length on a URI — no formal syntactic ceiling in the standard itself. |

Summary: 5 hollow · 2 partial.

---

### Grounding (`./assay examples/url-length/claim.txt -evidence -md -max-claims 15`)

Run: 2026-06-01, model `claude-sonnet-4-6`, 2 raw claims (no decomposition in evidence mode), 2 web searches, est $0.2052.

| # | Claim | Verdict | Finding | Sources |
|---|---|---|---|---|
| 1 | The maximum length of a URL is 2048 characters, a hard limit defined by the HTTP specification; any URL longer than that is invalid and will be rejected. | **refuted** | The HTTP specification imposes no hard limit on URL length whatsoever; the 2,048-character figure originates from Internet Explorer's path-length limit and the XML Sitemap standard, not from HTTP — and modern browsers routinely handle URLs far longer than 2,048 characters. | RFC 2616 (IETF); SISTRIX "URL Length"; Microsoft Support "Maximum URL length is 2,083 characters in Internet Explorer"; codegenes.net; GitHub puma/puma #2134 |
| 2 | The HTTP specification places no limit on URL length, so in practice a URL can be any length you want. | **mixed** | The first half is true — the HTTP specification imposes no hard limit — but the second half is false: browsers, web servers, CDNs, and proxies all enforce their own limits, making arbitrarily long URLs unreliable or outright rejected. | SISTRIX; urleditor.online; DevGex; urlencodedecode.com; GeeksforGeeks |

Summary: 1 refuted · 1 mixed.

**Reading the two tables together:** Substance catches the hidden premise
(`implementation = specification`) and the universal overreach (`any URL … is invalid`). Grounding
confirms: the spec claim is refuted by evidence, and the clever counter-claim
(`any length you want`) comes back mixed — the spec part survives, the practice part doesn't. The
armchair gets you to the right place on the spec; you still need the retrieval to find out what
deployed systems actually enforce.
