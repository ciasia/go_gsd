package pdf

import (
	"bytes"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_gsd/view"

	"log"
)

type PdfHandler struct {
	HandlerConfig  *PdfHandlerConfig
	TemplateWriter *view.TemplateWriter
	Binary         string
}

type PdfHandlerConfig struct {
	Templates map[string]view.TemplateConfig `json:"templates"`
}

func GetPdfHandler(binary string, handlerConfig *PdfHandlerConfig, templateWriter *view.TemplateWriter) (*PdfHandler, error) {
	eh := PdfHandler{
		HandlerConfig:  handlerConfig,
		TemplateWriter: templateWriter,
		Binary:         binary,
	}

	return &eh, nil
}

func (h *PdfHandler) Preview(requestTorch *torch.Request) {
	functionName := ""
	reportName := ""
	var id uint64

	err := requestTorch.UrlMatch(&functionName, &reportName, &id)
	if err != nil {
		log.Println(err)
		return
	}
	w := requestTorch.GetWriter()
	w.Header().Add("content-type", "text/html")
	reportConfig, ok := h.HandlerConfig.Templates[reportName]
	if !ok {
		log.Println("Template not found")
		return
	}
	err = h.TemplateWriter.Write(w, requestTorch, &reportConfig, id)
	if err != nil {
		log.Println(err)
		return
	}

}

func (h *PdfHandler) GetPdf(requestTorch *torch.Request) {
	functionName := ""
	reportName := ""
	var id uint64

	err := requestTorch.UrlMatch(&functionName, &reportName, &id)
	if err != nil {
		log.Println(err)
		return
	}
	w := bytes.Buffer{}
	reportConfig, ok := h.HandlerConfig.Templates[reportName]
	if !ok {
		log.Println("Template not found")
		return
	}
	err = h.TemplateWriter.Write(&w, requestTorch, &reportConfig, id)
	if err != nil {
		log.Println(err)
		return
	}
	r := bytes.NewReader(w.Bytes())
	DoPdf(h.Binary, r, requestTorch.GetWriter())
}