from html.parser import HTMLParser
from urllib.request import urlopen
from urllib import parse
class LinkParser(HTMLParser):
   def handle_starttag(self,tag,attrs):
      if tag=='a':
	     for(key,value)in attrs:
		    if key=='href':
			   newurl=parse.urljoin(self.baseurl,value)
			   self.links=self.links+[newurl]
	def getLinks(self,url):
	   self.links=[]
	   self.baseUrl=url
	   response=urlopen(url)
	   if response.getheader('Content-type')=='text/html':
	       htmlBytes=response.read()
		   htmlString=htmlBytes.decode("utf-8")
		   self.feed(htmlString)
		   return htmlString,self.links
		else:
		   return" ",[]
	def spider(url,word,maxPages):
		pagesToVisit=[url]
		numberVisited=0
		foundWord=False
		while numberVisited<maxPages and pagesToVisit !=[] and not foundWord : numberVisited=numberVisited+1
		   url pagesToVisit[0]
		   pagesToVisit=pagesToVisit[1:]
		   try:
				print(numberVisited,"visiting:",url)
				parser=LinkParser()
				data,Links=parser.getLinks(url)
				if data.find(word)>-1:
					foundWord=True
				pagesToVisit=pagestoVisit+links
				print("**success**")
			except:
				print("**Failed**")
		if foundWord:
			print("The Word",word,"was found at",url)
		else:
			print("word never found")
