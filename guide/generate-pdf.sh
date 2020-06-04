echo Generating PDF...
asciidoctor-pdf -a pdf-style=styles/pdf-theme.yml -a pdf-fontsdir=styles/fonts/ -D /documents/output index.adoc

echo Rename output PDF...
mv /documents/output/index.pdf /documents/output/seed-user-guide.pdf