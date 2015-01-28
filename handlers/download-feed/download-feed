#!/usr/bin/env python
import argparse
import bson
import cleanurl
import datetime
import feedparser
import pymongo
import time
import urllib.error

def process_feed(db, feed):
	etag = feed.get('etag', None)
	modified = feed.get('lastmodified', None)
	doc = feedparser.parse(feed['url'], etag=etag, modified=modified)

	print(feed['_id'], 'STATUS', doc['status'])

	update = {
		'bozo':         doc['bozo'],
		'etag':         doc.get('etag', None),
		'lastmodified': doc.get('modified', None),
		'laststatus':   doc['status'],
		'lastdownload': datetime.datetime.utcnow(),
		'nextdownload': datetime.datetime.utcnow() + datetime.timedelta(hours=1)
	}
	if 'updated' in doc:
		t = datetime.datetime.utcfromtimestamp(time.mktime(doc['updated_parsed']))
		update['updated'] = t
	if doc['status'] == 301:
		update['originalurl'] = feed['url']
		update['url'] = doc['href']

	db.feeds.update({ '_id': feed['_id'] }, { '$set': update })

	if doc['status'] == 410 or doc['status'] == 304:
		return

	for entry in doc.entries:
		try:
			url, req = cleanurl.clean(entry['link'])
		except urllib.error.HTTPError as err:
			print(feed['_id'], 'ERROR', err, entry['link'])
			db.article_errors.insert({
				'url':    entry['link'],
				'code':   err.code,
				'reason': err.reason
			})
			continue

		if db.articles.find_one({ 'url': url }) is not None:
			continue

		published = entry.get('published_parsed', None)
		if published is not None:
			published = datetime.datetime.utcfromtimestamp(time.mktime(published))

		html = req.read()

		article = {
			'_id':       bson.objectid.ObjectId(),
			'feedid':    feed['_id'],
			'pubid':     feed['pubid'],
			'author':    entry.get('author', None),
			'published': published,
			'title':     entry.get('title', None),
			'url':       url,
			'html':      html,
			'entry':     {
				'author':    entry.get('author', None),
				'published': published,
				'title':     entry.get('title', None),
				'url':       entry['link']
			}
		}
		
		print(feed['_id'], 'ARTICLE', article['_id'])
		db.articles.insert(article)


def main():
	parser = argparse.ArgumentParser(description='Download feed and push new URLs into next queue')
	parser.add_argument('-mongo', help='MongoDB connection string', default='mongodb://localhost:27017/ocular8')
	parser.add_argument('id', help='Feed ObjectId', nargs='+')
	args = parser.parse_args()

	db = pymongo.MongoClient(args.mongo).get_default_database()
	db.articles.ensure_index('url', unique=True)

	bson_ids = [ bson.objectid.ObjectId(id) for id in args.id ]
	feeds = db.feeds.find({ '_id': { '$in': bson_ids } })

	for feed in feeds:
		process_feed(db, feed)

if __name__ == '__main__':
	main()
