PHONY=all clean

all: DesignDocument.pdf

clean:
	rm -rf latex_out/
	rm DesignDocument.pdf

DesignDocument.pdf: designDocument.tex
	mkdir -p ./latex_out/
	pdflatex --output-dir=latex_out designDocument.tex
	pdflatex --output-dir=latex_out designDocument.tex
	pdflatex --output-dir=latex_out designDocument.tex
	cp latex_out/designDocument.pdf DesignDocument.pdf
