package web

import "embed"

// FS embeds the static web assets
//
//go:embed index.html style.css app.js
var FS embed.FS
