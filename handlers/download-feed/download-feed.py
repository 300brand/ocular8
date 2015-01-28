#!/usr/bin/env python
import argparse
import bson
import feedparser
import pymongo
import urlclean

def process_feed(db, feed):
	doc = feedparser.parse(feed['url'], etag=feed['etag'], modified=feed['lastmodified'])

	update = {
		'bozo':         doc['bozo'],
		'etag':         doc['etag'],
		'lastmodified': doc['modified'],
		'laststatus':   doc['status'],
		'updated':      doc['updated_parsed']
	}
	db.feeds.update({ '_id': feed['_id'] }, { '$set': update })

	for entry in feed.entries:
		article = {
			'_id':       bson.objectid.ObjectId(),
			'feedid':    feed['_id'],
			'pubid':     feed['pubid'],
			'author':    entry['author'],
			'published': entry['published_parsed'],
			'title':     entry['title'],
			'url':       urlclean.unshorten(entry['link']),
			'entry':     {
				'author':    entry['author'],
				'published': entry['published_parsed'],
				'url':       entry['link']
			}
		}
		article_exists = (db.articles.find_one({ 'url': article['url'] }) == None)
		if not article_exists:
			db.articles.insert(article)

def main():
	parser = argparse.ArgumentParser(description='Download feed and push new URLs into next queue')

	parser.add_argument('-mongo', help='MongoDB connection string', default='mongodb://localhost:27017/ocular8')
	parser.add_argument('id', help='Feed ObjectId', nargs='+')

	args = parser.parse_args()
	print(args.id)

	client = pymongo.MongoClient(args.mongo)

	db = client.get_default_database()

	bson_ids = [ bson.objectid.ObjectId(id) for id in args.id ]
	feeds = db.feeds.find({ '_id': { '$in': bson_ids } })

	for feed in feeds:
		process_feed(db, feed)

if __name__ == '__main__':
	main()
