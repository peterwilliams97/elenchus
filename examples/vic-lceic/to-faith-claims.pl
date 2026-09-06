#!/usr/bin/env perl
# to-faith-claims.pl report-summary.md claims-machine.txt > claims-faith.txt
#
# Turns the pipe-delimited claims-machine.txt into the one-claim-per-line format assay already reads,
# prefixing each line with "<id>\t<path>\t" so the tree renderer can key claims to the report's own
# headings. path = "<chap>=<chapter heading>/<ref>=<section heading>", both headings lifted from
# report-summary.md (data, not code). Faithfulness route only — route= is dropped, no dispatch.
use strict;
use warnings;
use utf8;
binmode STDOUT, ':encoding(UTF-8)';
my ($summary, $claims) = @ARGV;
my (%chap, %sect);
open my $s, '<:encoding(UTF-8)', $summary or die "open $summary: $!";
while (<$s>) {
    $chap{$1} = $2 if /^## Chapter (\d+) [—-] (.+?)\s*$/;
    $sect{$1} = $2 if /§(\d+\.\d+\.\d+) ([^):]+?)\s*[):]/ and !exists $sect{$1};
}
close $s;
open my $c, '<:encoding(UTF-8)', $claims or die "open $claims: $!";
while (<$c>) {
    next if /^#/ or /^\s*$/;
    my @f = split /\|/;
    next if @f < 4;
    (my $id = $f[0]) =~ s/\s+//g;
    my ($ref) = $f[2] =~ /§(\d+\.\d+(?:\.\d+)?)/;
    next unless $ref;
    (my $claim = $f[3]) =~ s/^\s+|\s+$//g;
    (my $chapn = $ref) =~ s/\..*//;
    my $clabel = $chap{$chapn} ? "Chapter $chapn — $chap{$chapn}" : "Chapter $chapn";
    my $slabel = $sect{$ref} // "\x{a7}$ref";
    print "$id\t$chapn=$clabel/$ref=$slabel\t$claim\n";
}
close $c;
