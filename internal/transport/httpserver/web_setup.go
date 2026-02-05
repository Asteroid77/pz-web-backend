package httpserver

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupStaticAndTemplates(r *gin.Engine, contentFS fs.FS) {
	assetsFS, _ := fs.Sub(contentFS, "template/assets")
	r.StaticFS("/assets", http.FS(assetsFS))

	tmpl := mustParseTemplates(contentFS)
	r.SetHTMLTemplate(tmpl)
}
