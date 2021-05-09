
gen-i18n-translation:
	langs=('en-US' 'zh-Hans'); \
	for lang in $${langs[@]}; \
	do \
		goi18n extract -format toml -outdir i18n/translation -sourceLanguage $$lang; \
	done
