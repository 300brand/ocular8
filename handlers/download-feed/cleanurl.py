#!/usr/bin/env python
import sys
import urllib.parse
import urllib.request

def clean(url):
	req = _resolve(url)
	return (_cleanquery(req.geturl()), req)

def _cleanquery(url):
	s = urllib.parse.urlsplit(url, allow_fragments=False)
	query = [ p for p in urllib.parse.parse_qsl(s.query) if _valid_param(p) ]
	return urllib.parse.urlunsplit([
		s.scheme,
		s.netloc,
		s.path,
		urllib.parse.urlencode(query),
		s.fragment
	])

def _resolve(url):
	return urllib.request.urlopen(url)

def _valid_param(p):
	k, v = p
	tests = [
		k.startswith('utm_'),
		k.startswith('fb'),
		k.startswith('atc'),
		'.99' in k,
		k in [
			'',
			'_r',
			'action_object_map',
			'action_ref_map',
			'action_type_map',
			'amp',
			'asrc',
			'beta',
			'CMP',
			'cmp',
			'cmpid',
			'comm_ref',
			'cpage',
			'dlvrit',
			'ex_cid',
			'f',
			'feedName',
			'feedType',
			'ft',
			'gplus',
			'kc',
			'lifehealth',
			'logvisit',
			'mbid',
			'ncid',
			'npu',
			'op',
			'rpc',
			's_cid',
			'sc',
			'source',
			'subj',
			'tag',
			'tc',
			'urw',
			'virtualBrandChannel',
		],
		v in [
			'rss',
		],
		k == "ref" and v == "25",
		k == "ana" and v == "RSS",
		k == "s" and v == "article_search",
		k == "attr" and v == "all",
	]
	return True not in tests

def main():
	[ print(i, clean(url)[0]) for i, url in enumerate(sys.argv[1:]) ]

if __name__ == "__main__":
	main()
