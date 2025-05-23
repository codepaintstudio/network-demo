package services

import (
	"errors"
	"server/internal/models"

	"gorm.io/gorm"
)

type ArticleService struct {
	db *gorm.DB
}

func NewArticleService(db *gorm.DB) *ArticleService {
	return &ArticleService{db: db}
}

// Create 创建帖子
func (s *ArticleService) Create(userId int, request *models.CreateArticleRequest) (*models.Article, error) {
	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, userId).Error; err != nil {
		return nil, err
	}

	article := models.Article{
		Title:   request.Title,
		Content: request.Content,
		UserId:  userId,
		User:    user,
		UserInfo: models.UserInfo{
			Id:       user.Id,
			Username: user.Username,
			Email:    user.Email,
		},
	}

	if err := s.db.Create(&article).Error; err != nil {
		return nil, err
	}

	// 填充用户信息
	article.UserInfo = models.UserInfo{
		Id:       article.User.Id,
		Username: article.User.Username,
		Email:    article.User.Email,
	}

	return &article, nil
}

// Update 更新帖子
func (s *ArticleService) Update(userId int, articleId int, request *models.UpdateArticleRequest) (*models.Article, error) {
	// 查询帖子
	var article models.Article
	if err := s.db.Preload("User").First(&article, articleId).Error; err != nil {
		return nil, err
	}

	// 检查帖子所有权
	if article.UserId != userId {
		return nil, errors.New("unauthorized to update this article")
	}

	// 更新帖子
	article.Title = request.Title
	article.Content = request.Content

	if err := s.db.Save(&article).Error; err != nil {
		return nil, err
	}

	// 填充用户信息
	article.UserInfo = models.UserInfo{
		Id:       article.User.Id,
		Username: article.User.Username,
		Email:    article.User.Email,
	}

	return &article, nil
}

// Delete 删除帖子
func (s *ArticleService) Delete(userId int, articleId int) error {
	var article models.Article
	if err := s.db.First(&article, articleId).Error; err != nil {
		return err
	}

	// 验证帖子所有者
	if article.UserId != userId {
		return errors.New("unauthorized to delete this article")
	}

	return s.db.Delete(&article).Error
}

// GetById 获取帖子详情
func (s *ArticleService) GetById(articleId int) (*models.Article, error) {
	var article models.Article
	if err := s.db.Preload("User").First(&article, articleId).Error; err != nil {
		return nil, err
	}

	// 填充用户信息（不含密码）
	article.UserInfo = models.UserInfo{
		Id:       article.User.Id,
		Username: article.User.Username,
		Email:    article.User.Email,
	}

	return &article, nil
}

// List 获取帖子列表
func (s *ArticleService) List(request *models.ArticleListRequest) (*models.ArticleListResponse, error) {
	var total int64
	var articles []models.Article

	offset := (request.Page - 1) * request.Size

	// 获取总数
	if err := s.db.Model(&models.Article{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// 获取分页数据
	if err := s.db.Preload("User").Order("created_at desc").Offset(offset).Limit(request.Size).Find(&articles).Error; err != nil {
		return nil, err
	}

	// 填充用户信息（不含密码）
	articleResponses := make([]models.Article, len(articles))
	for i, article := range articles {
		articleResponses[i] = article
		articleResponses[i].UserInfo = models.UserInfo{
			Id:       article.User.Id,
			Username: article.User.Username,
			Email:    article.User.Email,
		}
	}

	return &models.ArticleListResponse{
		Articles: articleResponses,
		Total:    int(total),
		Page:     request.Page,
		Size:     request.Size,
	}, nil
}
