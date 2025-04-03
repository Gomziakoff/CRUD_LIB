package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Gomziakoff/CRUD_LIB/internal/domain"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Books interface {
	Create(ctx context.Context, book domain.Book) error
	GetByID(ctx context.Context, id int64) (domain.Book, error)
	GetAll(ctx context.Context) ([]domain.Book, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, id int64, inp domain.UpdateBookInput) error
}

type Handler struct {
	booksService Books
}

func NewHandler(books Books) *Handler {
	return &Handler{
		booksService: books,
	}
}

// @title Books API
// @version 1.0
// @description CRUD API для управления книгами
// @host localhost:8080
// @BasePath /

// InitRouter инициализирует роутер
func (h *Handler) InitRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	// Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	books := r.PathPrefix("/books").Subrouter()
	{
		books.HandleFunc("", h.createBook).Methods(http.MethodPost)
		books.HandleFunc("", h.getAllBooks).Methods(http.MethodGet)
		books.HandleFunc("/{id:[0-9]+}", h.getBookByID).Methods(http.MethodGet)
		books.HandleFunc("/{id:[0-9]+}", h.deleteBook).Methods(http.MethodDelete)
		books.HandleFunc("/{id:[0-9]+}", h.updateBook).Methods(http.MethodPut)
	}
	return r
}

// GetBookByID godoc
// @Summary Получить книгу по ID
// @Description Получить книгу по её идентификатору
// @Tags books
// @Accept  json
// @Produce  json
// @Param id path int true "ID книги"
// @Success 200 {object} domain.Book
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [get]
func (h *Handler) getBookByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIdFromRequest(r)
	if err != nil {
		log.Println("getBookByID() error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	book, err := h.booksService.GetByID(context.TODO(), id)
	if err != nil {
		if errors.Is(err, domain.ErrBookNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Println("getBookById() error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(book)
	if err != nil {
		log.Println("getBookByID() error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}

// CreateBook godoc
// @Summary Создать новую книгу
// @Description Создание новой книги в системе
// @Tags books
// @Accept  json
// @Produce  json
// @Param input body domain.Book true "Данные книги"
// @Success 201
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books [post]
func (h *Handler) createBook(w http.ResponseWriter, r *http.Request) {
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"handler": "createBook",
			"problem": "reading request",
		}).Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var book domain.Book
	if err = json.Unmarshal(reqBytes, &book); err != nil {
		log.WithFields(log.Fields{
			"handler": "createBook",
			"problem": "unmarshaling request",
		}).Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.booksService.Create(context.TODO(), book)
	if err != nil {
		log.WithFields(log.Fields{
			"handler": "createBook",
			"problem": "service error",
		}).Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info(fmt.Sprintf("Created book '%s'", book.Title))
	w.WriteHeader(http.StatusCreated)
}

// DeleteBook godoc
// @Summary Удалить книгу
// @Description Удаление книги по её ID
// @Tags books
// @Accept  json
// @Produce  json
// @Param id path int true "ID книги"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [delete]
func (h *Handler) deleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := getIdFromRequest(r)
	if err != nil {
		log.Println("deleteBook() error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.booksService.Delete(context.TODO(), id)
	if err != nil {
		log.Println("deleteBook() error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetAllBooks godoc
// @Summary Получить все книги
// @Description Получение списка всех книг
// @Tags books
// @Accept  json
// @Produce  json
// @Success 200 {array} domain.Book
// @Failure 500 {object} ErrorResponse
// @Router /books [get]
func (h *Handler) getAllBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.booksService.GetAll(context.TODO())
	if err != nil {
		log.Println("getAllBooks() error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(books)
	if err != nil {
		log.Println("getAllBooks() error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}

// UpdateBook godoc
// @Summary Обновить книгу
// @Description Обновление данных книги
// @Tags books
// @Accept  json
// @Produce  json
// @Param id path int true "ID книги"
// @Param input body domain.UpdateBookInput true "Обновляемые поля"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [put]
func (h *Handler) updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := getIdFromRequest(r)
	if err != nil {
		log.Println("error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var inp domain.UpdateBookInput
	if err = json.Unmarshal(reqBytes, &inp); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.booksService.Update(context.TODO(), id, inp)
	if err != nil {
		log.Println("error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ErrorResponse структура для ошибок
type ErrorResponse struct {
	Error string `json:"error"`
}

func getIdFromRequest(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return 0, err
	}

	if id == 0 {
		return 0, errors.New("id can't be 0")
	}

	return id, nil
}
