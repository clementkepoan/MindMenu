package main

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
    // Health check endpoint
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
        })
    })

    // Example endpoints from handlers.go
    r.GET("/example", GetExample)
    r.POST("/example", PostExample)

    // Simple file upload endpoint (without Supabase storage for now)
    r.POST("/upload", func(c *gin.Context) {
        file, err := c.FormFile("file")
        if err != nil {
            c.JSON(400, gin.H{"error": "No file provided"})
            return
        }

        c.JSON(200, gin.H{
            "message":  "File received successfully",
            "filename": file.Filename,
            "size":     file.Size,
            "note":     "Storage implementation can be added later",
        })
    })
}