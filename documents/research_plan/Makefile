PHONY=all clean

all: Research_plan.pdf 

clean:
	rm -rf latex_out/
	rm Research_plan.pdf 

Research_plan.pdf: research_plan.tex
	mkdir -p ./latex_out/
	pdflatex --output-dir=latex_out research_plan.tex
	pdflatex --output-dir=latex_out research_plan.tex
	pdflatex --output-dir=latex_out research_plan.tex
	cp latex_out/research_plan.pdf Research_plan.pdf 
