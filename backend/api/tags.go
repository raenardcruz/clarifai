package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"note-taker/backend/models"
)

type TagSummary struct {
	ID         uint    `json:"id"`
	ParentID   *uint   `json:"parent_id"`
	Name       string  `json:"name"`
	Color      string  `json:"color"`
	Path       string  `json:"path"`
	ShareToken *string `json:"share_token,omitempty"`
}

type TagNode struct {
	ID          uint      `json:"id"`
	ParentID    *uint     `json:"parent_id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	Path        string    `json:"path"`
	ShareToken  *string   `json:"share_token,omitempty"`
	DirectCount int       `json:"direct_count"`
	TotalCount  int       `json:"total_count"`
	Children    []TagNode `json:"children"`
}

type NullableUint struct {
	Set   bool
	Value *uint
}

func (n *NullableUint) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		n.Value = nil
		return nil
	}
	var v uint
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	n.Value = &v
	return nil
}

type CreateTagRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID *uint  `json:"parent_id"`
	Color    string `json:"color"`
}

type UpdateTagRequest struct {
	Name     *string      `json:"name"`
	Color    *string      `json:"color"`
	ParentID NullableUint `json:"parent_id"`
}

func SetupTagsRoutes(router *gin.RouterGroup) {
	// Public read-only tag sharing route (no auth required)
	router.GET("/share/:token", getSharedTag)

	authGroup := router.Group("")
	authGroup.Use(AuthMiddleware(), RequireApprovedUser())
	{
		authGroup.GET("", getTags)
		authGroup.POST("", createTag)
		authGroup.PATCH("/:id", updateTag)
		authGroup.DELETE("/:id", deleteTag)
		authGroup.POST("/:id/share", shareTag)
		authGroup.DELETE("/:id/share", unshareTag)
	}
}

func loadUserTagGraph(db *gorm.DB, userID uint) ([]models.Tag, map[uint]models.Tag, map[uint][]uint) {
	var tags []models.Tag
	db.Where("user_id = ?", userID).Order("LOWER(name) ASC").Find(&tags)

	byID := make(map[uint]models.Tag, len(tags))
	childrenMap := make(map[uint][]uint)
	for _, t := range tags {
		byID[t.ID] = t
		if t.ParentID != nil {
			childrenMap[*t.ParentID] = append(childrenMap[*t.ParentID], t.ID)
		}
	}
	return tags, byID, childrenMap
}

func buildTagPath(tagID uint, byID map[uint]models.Tag) string {
	var parts []string
	visited := make(map[uint]bool)
	currID := &tagID
	for currID != nil {
		if visited[*currID] {
			break
		}
		visited[*currID] = true
		t, ok := byID[*currID]
		if !ok {
			break
		}
		parts = append([]string{t.Name}, parts...)
		currID = t.ParentID
	}
	return strings.Join(parts, " / ")
}

func getDescendantTagIDs(tagID uint, childrenMap map[uint][]uint) []uint {
	result := []uint{tagID}
	visited := map[uint]bool{tagID: true}
	queue := []uint{tagID}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, childID := range childrenMap[curr] {
			if !visited[childID] {
				visited[childID] = true
				result = append(result, childID)
				queue = append(queue, childID)
			}
		}
	}
	return result
}

func buildTagSummaries(tags []models.Tag, byID map[uint]models.Tag) []TagSummary {
	summaries := make([]TagSummary, 0, len(tags))
	for _, t := range tags {
		summaries = append(summaries, TagSummary{
			ID:       t.ID,
			ParentID: t.ParentID,
			Name:     t.Name,
			Color:    t.Color,
			Path:     buildTagPath(t.ID, byID),
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		return strings.ToLower(summaries[i].Path) < strings.ToLower(summaries[j].Path)
	})
	return summaries
}

func checkSiblingDuplicate(db *gorm.DB, userID uint, parentID *uint, name string, excludeID *uint) (bool, error) {
	query := db.Model(&models.Tag{}).Where("user_id = ? AND LOWER(name) = LOWER(?)", userID, strings.TrimSpace(name))
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func getTags(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	db := models.GetDB()
	tags, byID, childrenMap := loadUserTagGraph(db, currentUser.ID)

	// Load recording_tags for this user's recordings to compute direct and rolled-up counts
	type recTagRow struct {
		TagID       uint   `gorm:"column:tag_id"`
		RecordingID string `gorm:"column:recording_id"`
	}
	var rows []recTagRow
	db.Table("recording_tags").
		Select("recording_tags.tag_id, recording_tags.recording_id").
		Joins("JOIN recordings ON recordings.id = recording_tags.recording_id").
		Where("recordings.user_id = ?", currentUser.ID).
		Scan(&rows)

	directRecsByTag := make(map[uint]map[string]bool)
	for _, row := range rows {
		if directRecsByTag[row.TagID] == nil {
			directRecsByTag[row.TagID] = make(map[string]bool)
		}
		directRecsByTag[row.TagID][row.RecordingID] = true
	}

	var untaggedCount int64
	db.Model(&models.Recording{}).
		Where("user_id = ? AND NOT EXISTS (SELECT 1 FROM recording_tags WHERE recording_tags.recording_id = recordings.id)", currentUser.ID).
		Count(&untaggedCount)

	// Compute recursive total recording set per tag
	var collectTotalRecs func(tagID uint, visited map[uint]bool) map[string]bool
	collectTotalRecs = func(tagID uint, visited map[uint]bool) map[string]bool {
		recs := make(map[string]bool)
		if visited[tagID] {
			return recs
		}
		visited[tagID] = true
		for recID := range directRecsByTag[tagID] {
			recs[recID] = true
		}
		for _, childID := range childrenMap[tagID] {
			for recID := range collectTotalRecs(childID, visited) {
				recs[recID] = true
			}
		}
		return recs
	}

	flatNodes := make([]TagNode, 0, len(tags))
	nodeByID := make(map[uint]TagNode, len(tags))
	for _, t := range tags {
		directCount := len(directRecsByTag[t.ID])
		totalCount := len(collectTotalRecs(t.ID, make(map[uint]bool)))
		node := TagNode{
			ID:          t.ID,
			ParentID:    t.ParentID,
			Name:        t.Name,
			Color:       t.Color,
			Path:        buildTagPath(t.ID, byID),
			ShareToken:  t.ShareToken,
			DirectCount: directCount,
			TotalCount:  totalCount,
			Children:    []TagNode{},
		}
		flatNodes = append(flatNodes, node)
		nodeByID[t.ID] = node
	}

	var buildTree func(parentID *uint) []TagNode
	buildTree = func(parentID *uint) []TagNode {
		var branch []TagNode
		for _, t := range tags {
			matches := (parentID == nil && t.ParentID == nil) || (parentID != nil && t.ParentID != nil && *parentID == *t.ParentID)
			if matches {
				node := nodeByID[t.ID]
				node.Children = buildTree(&t.ID)
				branch = append(branch, node)
			}
		}
		return branch
	}

	tree := buildTree(nil)
	if tree == nil {
		tree = []TagNode{}
	}

	c.JSON(http.StatusOK, gin.H{
		"tags":           flatNodes,
		"tree":           tree,
		"untagged_count": untaggedCount,
	})
}

func createTag(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Tag name cannot be empty"})
		return
	}

	db := models.GetDB()
	if req.ParentID != nil {
		var parent models.Tag
		if err := db.Where("id = ? AND user_id = ?", *req.ParentID, currentUser.ID).First(&parent).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Parent tag not found"})
			return
		}
	}

	dup, err := checkSiblingDuplicate(db, currentUser.ID, req.ParentID, name, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to validate tag"})
		return
	}
	if dup {
		c.JSON(http.StatusConflict, gin.H{"detail": "A tag with this name already exists at this level"})
		return
	}

	color := strings.TrimSpace(req.Color)
	if color == "" {
		color = "#3b82f6"
	}

	tag := models.Tag{
		UserID:   currentUser.ID,
		ParentID: req.ParentID,
		Name:     name,
		Color:    color,
	}

	if err := db.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create tag"})
		return
	}

	_, byID, _ := loadUserTagGraph(db, currentUser.ID)
	c.JSON(http.StatusCreated, TagNode{
		ID:       tag.ID,
		ParentID: tag.ParentID,
		Name:     tag.Name,
		Color:    tag.Color,
		Path:     buildTagPath(tag.ID, byID),
		Children: []TagNode{},
	})
}

func updateTag(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	tagID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid tag ID"})
		return
	}
	tagID := uint(tagID64)

	db := models.GetDB()
	var tag models.Tag
	if err := db.Where("id = ? AND user_id = ?", tagID, currentUser.ID).First(&tag).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Tag not found"})
		return
	}

	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	targetParentID := tag.ParentID
	if req.ParentID.Set {
		targetParentID = req.ParentID.Value
		if targetParentID != nil {
			if *targetParentID == tag.ID {
				c.JSON(http.StatusBadRequest, gin.H{"detail": "A tag cannot be its own parent"})
				return
			}
			var parent models.Tag
			if err := db.Where("id = ? AND user_id = ?", *targetParentID, currentUser.ID).First(&parent).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"detail": "Target parent tag not found"})
				return
			}
			_, _, childrenMap := loadUserTagGraph(db, currentUser.ID)
			descendants := getDescendantTagIDs(tag.ID, childrenMap)
			for _, dID := range descendants {
				if dID == *targetParentID {
					c.JSON(http.StatusBadRequest, gin.H{"detail": "Cannot move a tag under one of its own descendants"})
					return
				}
			}
		}
	}

	targetName := tag.Name
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Tag name cannot be empty"})
			return
		}
		targetName = trimmed
	}

	dup, err := checkSiblingDuplicate(db, currentUser.ID, targetParentID, targetName, &tag.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to validate tag"})
		return
	}
	if dup {
		c.JSON(http.StatusConflict, gin.H{"detail": "A tag with this name already exists at the target level"})
		return
	}

	tag.Name = targetName
	tag.ParentID = targetParentID
	if req.Color != nil && strings.TrimSpace(*req.Color) != "" {
		tag.Color = strings.TrimSpace(*req.Color)
	}

	if err := db.Save(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to update tag"})
		return
	}

	_, byID, _ := loadUserTagGraph(db, currentUser.ID)
	c.JSON(http.StatusOK, TagSummary{
		ID:         tag.ID,
		ParentID:   tag.ParentID,
		Name:       tag.Name,
		Color:      tag.Color,
		Path:       buildTagPath(tag.ID, byID),
		ShareToken: tag.ShareToken,
	})
}

func deleteTag(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	tagID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid tag ID"})
		return
	}
	tagID := uint(tagID64)
	mode := c.DefaultQuery("mode", "reparent")

	db := models.GetDB()
	var tag models.Tag
	if err := db.Where("id = ? AND user_id = ?", tagID, currentUser.ID).First(&tag).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Tag not found"})
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if mode == "cascade" {
			_, _, childrenMap := loadUserTagGraph(tx, currentUser.ID)
			toDelete := getDescendantTagIDs(tag.ID, childrenMap)
			if err := tx.Exec("DELETE FROM recording_tags WHERE tag_id IN ?", toDelete).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ? AND user_id = ?", toDelete, currentUser.ID).Delete(&models.Tag{}).Error; err != nil {
				return err
			}
		} else {
			// Reparent direct children to the deleted tag's parent
			if err := tx.Model(&models.Tag{}).
				Where("parent_id = ? AND user_id = ?", tag.ID, currentUser.ID).
				Update("parent_id", tag.ParentID).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM recording_tags WHERE tag_id = ?", tag.ID).Error; err != nil {
				return err
			}
			if err := tx.Delete(&tag).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to delete tag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted successfully"})
}

// buildSubtreeTagPath builds a tag path stopped at rootTagID so ancestor tags outside
// the shared subtree are never exposed to guests.
func buildSubtreeTagPath(tagID uint, rootTagID uint, byID map[uint]models.Tag) string {
	var parts []string
	visited := make(map[uint]bool)
	currID := &tagID
	for currID != nil {
		if visited[*currID] {
			break
		}
		visited[*currID] = true
		t, ok := byID[*currID]
		if !ok {
			break
		}
		parts = append([]string{t.Name}, parts...)
		if t.ID == rootTagID {
			break
		}
		currID = t.ParentID
	}
	return strings.Join(parts, " / ")
}

func shareTag(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	tagID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid tag ID"})
		return
	}
	tagID := uint(tagID64)

	db := models.GetDB()
	var tag models.Tag
	if err := db.Where("id = ? AND user_id = ?", tagID, currentUser.ID).First(&tag).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Tag not found"})
		return
	}

	if tag.ShareToken == nil || strings.TrimSpace(*tag.ShareToken) == "" {
		token := uuid.New().String()
		if err := db.Model(&tag).Update("share_token", token).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to generate share link"})
			return
		}
		tag.ShareToken = &token
	}

	c.JSON(http.StatusOK, gin.H{
		"tag_id":      tag.ID,
		"share_token": *tag.ShareToken,
	})
}

func unshareTag(c *gin.Context) {
	currentUser, err := getContextUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	tagID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid tag ID"})
		return
	}
	tagID := uint(tagID64)

	db := models.GetDB()
	var tag models.Tag
	if err := db.Where("id = ? AND user_id = ?", tagID, currentUser.ID).First(&tag).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Tag not found"})
		return
	}

	if err := db.Model(&tag).Update("share_token", nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to revoke share link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag share link revoked"})
}

func getSharedTag(c *gin.Context) {
	token := strings.TrimSpace(c.Param("token"))
	if token == "" {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Shared tag not found"})
		return
	}

	db := models.GetDB()
	var tag models.Tag
	if err := db.Where("share_token = ?", token).First(&tag).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Shared tag not found or link has been revoked"})
		return
	}

	_, byID, childrenMap := loadUserTagGraph(db, tag.UserID)
	subtreeIDs := getDescendantTagIDs(tag.ID, childrenMap)
	subtreeSet := make(map[uint]bool, len(subtreeIDs))
	for _, id := range subtreeIDs {
		subtreeSet[id] = true
	}

	var recordings []models.Recording
	db.Preload("Tags").
		Where("user_id = ? AND EXISTS (SELECT 1 FROM recording_tags WHERE recording_tags.recording_id = recordings.id AND recording_tags.tag_id IN ?)", tag.UserID, subtreeIDs).
		Order("created_at DESC").
		Find(&recordings)

	subtags := make([]TagSummary, 0, len(subtreeIDs)-1)
	for _, id := range subtreeIDs {
		if id == tag.ID {
			continue
		}
		if t, ok := byID[id]; ok {
			subtags = append(subtags, TagSummary{
				ID:       t.ID,
				ParentID: t.ParentID,
				Name:     t.Name,
				Color:    t.Color,
				Path:     buildSubtreeTagPath(t.ID, tag.ID, byID),
			})
		}
	}
	sort.Slice(subtags, func(i, j int) bool {
		return strings.ToLower(subtags[i].Path) < strings.ToLower(subtags[j].Path)
	})

	sharedRecordings := make([]gin.H, 0, len(recordings))
	for _, rec := range recordings {
		recTags := make([]TagSummary, 0)
		for _, t := range rec.Tags {
			if subtreeSet[t.ID] {
				recTags = append(recTags, TagSummary{
					ID:       t.ID,
					ParentID: t.ParentID,
					Name:     t.Name,
					Color:    t.Color,
					Path:     buildSubtreeTagPath(t.ID, tag.ID, byID),
				})
			}
		}
		sort.Slice(recTags, func(i, j int) bool {
			return strings.ToLower(recTags[i].Path) < strings.ToLower(recTags[j].Path)
		})

		sharedRecordings = append(sharedRecordings, gin.H{
			"id":         rec.ID,
			"title":      rec.Title,
			"status":     rec.Status,
			"created_at": rec.CreatedAt.Format(time.RFC3339),
			"duration":   rec.Duration,
			"summary_md": rec.SummaryMD,
			"is_medical": rec.IsMedical,
			"tags":       recTags,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tag": TagSummary{
			ID:    tag.ID,
			Name:  tag.Name,
			Color: tag.Color,
			Path:  tag.Name,
		},
		"subtags":    subtags,
		"recordings": sharedRecordings,
	})
}
