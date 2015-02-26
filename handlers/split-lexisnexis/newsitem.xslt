<?xml version="1.0" encoding="UTF-8"?>

<xsl:stylesheet
	version="1.0"
	xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
	xmlns:co="http://www.lexis-nexis.com/co"
	xmlns:codir="http://www.lexis-nexis.com/codir"
	xmlns:coprof="http://www.lexis-nexis.com/coprof"
	xmlns:fin="http://www.lexis-nexis.com/fin"
	xmlns:indrep="http://www.lexis-nexis.com/indrep"
	xmlns:legidx="http://www.lexis-nexis.com/legidx"
	xmlns:lnci="http://www.lexis-nexis.com/lnci"
	xmlns:lncle="http://www.lexis-nexis.com/lncle"
	xmlns:lnclx="http://www.lexis-nexis.com/lnclx"
	xmlns:lndel="http://www.lexis-nexis.com/lndel"
	xmlns:lndocmeta="http://www.lexis-nexis.com/lndocmeta"
	xmlns:lngntxt="http://www.lexis-nexis.com/lngntxt"
	xmlns:lngt="http://www.lexis-nexis.com/lngt"
	xmlns:lnlit="http://www.lexis-nexis.com/lnlit"
	xmlns:lnsys="http://www.lexis-nexis.com/lnsys"
	xmlns:lnv="http://www.lexis-nexis.com/lnv"
	xmlns:lnvni="http://www.lexis-nexis.com/lnvni"
	xmlns:lnvx="http://www.lexis-nexis.com/lnvx"
	xmlns:lnvxe="http://www.lexis-nexis.com/lnvxe"
	xmlns:m="http://www.w3.org/1999/mathml"
	xmlns:m-a="http://www.lexis-nexis.com/m-a"
	xmlns:nitf="urn:nitf:iptc.org.20010418.NITF"
	xmlns:pat="http://www.lexis-nexis.com/pat"
	xmlns:peoref="http://www.lexis-nexis.com/peoref"
	xmlns:person="http://www.lexis-nexis.com/person"
	xmlns:research="http://www.lexis-nexis.com/research"
	xmlns:sa="http://www.lexis-nexis.com/sa"
	xmlns:sec="http://www.lexis-nexis.com/sec"
	xmlns:secfile="http://www.lexis-nexis.com/secfile"
	xmlns:stock="http://www.lexis-nexis.com/stock"
	exclude-result-prefixes="co codir coprof fin indrep legidx lnci lncle lnclx lndel lndocmeta lngntxt lngt lnlit lnsys lnv lnvni lnvx lnvxe m-a m nitf pat peoref person research sa sec secfile stock"
	>

	<xsl:output method="xml" encoding="UTF-8" omit-xml-declaration="yes" indent="yes" standalone="yes" />

	<xsl:template match="lnvxe:url">
		<a>
			<xsl:attribute name="href">
				<xsl:value-of select="remotelink/@href" />
			</xsl:attribute>
			<xsl:value-of select="remotelink" />
		</a>
	</xsl:template>

	<xsl:template match="nl">
		<br />
	</xsl:template>

	<xsl:template match="node()">
		<xsl:copy>
			<xsl:copy-of select="@*" />
			<xsl:apply-templates select="*|text()" />
		</xsl:copy>
	</xsl:template>

	<xsl:template match="/NEWSITEM">
		<html>
			<xsl:attribute name="lang">
				<xsl:value-of select="lnv:LANGUAGE/lnvxe:lang.english/@iso639-1" />
			</xsl:attribute>
			<xsl:attribute name="itemtype">http://schema.org/NewsArticle</xsl:attribute>
			<head>
				<meta charset="UTF-8" />
				<script type="application/xml">
					<!--
						Wish this could be applicaton/json, but writing JSON
						with XML is probably not going to end well for anyone
					-->
					<lexisnexis>
						<lni>
							<xsl:value-of select="lndocmeta:docinfo/lndocmeta:lnlni/@lnlni" />
						</lni>
						<dpsi>
							<xsl:value-of select="lndocmeta:docinfo/lndocmeta:dpsi/@lndpsi" />
						</dpsi>
						<pub>
							<xsl:value-of select="lnv:PUB" />
						</pub>
						<pubtype>
							<xsl:value-of select="lnv:PUBLICATION-TYPE/lnvxe:desc" />
						</pubtype>
						<url>
							<xsl:value-of select="lnv:URL-SEG/lnvxe:url/remotelink/@href" />
						</url>
						<language>
							<xsl:attribute name="name">
								<xsl:value-of select="lnv:LANGUAGE/lnvxe:lang.english" />
							</xsl:attribute>
							<xsl:value-of select="lnv:LANGUAGE/lnvxe:lang.english/@iso639-1" />
						</language>
						<section>
							<xsl:value-of select="lnv:SECTION-INFO/lnvxe:position.section" />
						</section>
					</lexisnexis>
				</script>
				<link rel="canonical">
					<xsl:attribute name="href">
						<xsl:value-of select="lnv:URL-SEG/lnvxe:url/remotelink/@href" />
					</xsl:attribute>
				</link>
				<meta name="headline">
					<xsl:attribute name="content">
						<xsl:value-of select="lnv:HEADLINE/lnvxe:hl1" />
					</xsl:attribute>
				</meta>
				<!-- Schema.org below -->
				<meta itemprop="printColumn">
					<xsl:attribute name="content">
						<xsl:value-of select="lnv:COLUMN" />
					</xsl:attribute>
				</meta>
				<meta itemprop="printEdition">
					<xsl:attribute name="content">
						<xsl:value-of select="lnv:EDITION/lnvxe:desc[0]" />
					</xsl:attribute>
				</meta>
				<meta itemprop="printPage">
					<xsl:attribute name="content">
						<xsl:value-of select="lnv:SECTION-INFO/lnvxe:position.sequence" />
					</xsl:attribute>
				</meta>
				<meta itemprop="printSection">
					<xsl:attribute name="content">
						<xsl:value-of select="lnv:SECTION-INFO/lnvxe:position.section" />
					</xsl:attribute>
				</meta>
				<meta itemprop="datePublished">
					<xsl:attribute name="content">
						<xsl:value-of select="lnv:PUB-DATE/lnvxe:date/@year" />
						<xsl:text>-</xsl:text>
						<xsl:value-of select="lnv:PUB-DATE/lnvxe:date/@month" />
						<xsl:text>-</xsl:text>
						<xsl:value-of select="lnv:PUB-DATE/lnvxe:date/@day" />
					</xsl:attribute>
				</meta>
				<meta itemprop="copyrightHolder">
					<xsl:attribute name="content">
						<xsl:value-of select="lnv:COPYRIGHT/lnv:PUB-COPYRIGHT/lnvxe:copyright.holder" />
					</xsl:attribute>
				</meta>
				<meta itemprop="copyrightYear">
					<xsl:attribute name="content">
						<xsl:value-of select="lnv:COPYRIGHT/lnv:PUB-COPYRIGHT/lnvxe:copyright.year/@year" />
					</xsl:attribute>
				</meta>
				<meta itemprop="wordCount">
					<xsl:attribute name="content">
						<xsl:value-of select="lnv:LENGTH" />
					</xsl:attribute>
				</meta>
				<meta itemprop="url">
					<xsl:attribute name="content">
						<xsl:value-of select="lnv:URL-SEG/lnvxe:url/remotelink/@href" />
					</xsl:attribute>
				</meta>
			</head>
			<body>
				<div itemprop="articleBody">
					<xsl:apply-templates select="lnv:REAL-LEAD/*" />
					<xsl:apply-templates select="lnv:BODY-1/*" />
					<xsl:apply-templates select="lnv:BODY-2/*" />
					<xsl:apply-templates select="lnv:BODY-3/*" />
					<xsl:apply-templates select="lnv:BODY-4/*" />
					<xsl:apply-templates select="lnv:BODY-5/*" />
					<xsl:apply-templates select="lnv:BODY-6/*" />
					<xsl:apply-templates select="lnv:BODY-7/*" />
					<xsl:apply-templates select="lnv:BODY-8/*" />
					<xsl:apply-templates select="lnv:BODY-9/*" />
					<xsl:apply-templates select="lnv:BODY-10/*" />
					<xsl:apply-templates select="lnv:BODY-11/*" />
					<xsl:apply-templates select="lnv:BODY-12/*" />
					<xsl:apply-templates select="lnv:BODY-13/*" />
					<xsl:apply-templates select="lnv:BODY-14/*" />
				</div>
			</body>
		</html>
	</xsl:template>

</xsl:stylesheet>
