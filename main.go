package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// 1. 데이터 모델 (동일)
type Student struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	BorrowedBooks []string `json:"borrowed_books"`
}

// 2. In-Memory Map 데이터베이스
var studentsDB *sql.DB

func main() {

	// 1번
	connStr := os.Getenv("DB_SOURCE")

	if connStr == "" {
		log.Println("DB_SOURCE 환경 변수가 설정되지 않았습니다. 로컬 개발용 개별 환경 변수를 읽습니다.")

		dbUser := os.Getenv("POSTGRES_USER")
		dbPass := os.Getenv("POSTGRES_PASSWORD")
		dbName := os.Getenv("POSTGRES_DB")
		dbHost := "localhost"
		dbPort := "5432"

		if dbUser == "" || dbPass == "" || dbName == "" {
			log.Fatal("로컬 개발용 환경 변수(POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB)가 설정되지 않았습니다. .env 파일을 참조하여 설정해주세요.")
		}

		connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPass, dbHost, dbPort, dbName)
	}
	var err error
	studentsDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("DB Connect 폴 생성 실패", err)
	}
	defer studentsDB.Close()

	if err = studentsDB.Ping(); err != nil {
		log.Fatal("DB 연결 실패:", err)
	}
	log.Println("데이터베이스 연결 성공")

	r := chi.NewRouter()
	r.Use(middleware.Logger) // 로깅 미들웨어

	// 3. 라우팅 (동일)
	r.Get("/students", GetAllStudents)
	r.Post("/students", CreateStudent)
	r.Get("/students/{id}", GetStudentByID)
	r.Put("/students/{id}", UpdateStudent)
	r.Delete("/students/{id}", DeleteStudent)

	// 4. 서버 시작 (동일)
	log.Println("서버 시작 (chi + map): http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

func GetAllStudents(w http.ResponseWriter, r *http.Request) {
	rows, err := studentsDB.QueryContext(r.Context(), "SELECT id, name, email, borrowed_books FROM students")
	if err != nil {
		log.Printf("DB 쿼리 실패 (%s): %v", "GetAllStudents", err)
		http.Error(w, "DB 조회 실패", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var studentList []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.ID, &s.Name, &s.Email, pq.Array(&s.BorrowedBooks)); err != nil {
			http.Error(w, "DB 스캔 실패", http.StatusInternalServerError)
			return
		}
		studentList = append(studentList, s)
	}

	jsonData, err := json.Marshal(studentList)
	if err != nil {
		log.Printf("DB 쿼리 실패 (%s): %v", "GetAllStudents", err)
		http.Error(w, "JSON 변환 실패", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func CreateStudent(w http.ResponseWriter, r *http.Request) {
	var newStudent Student
	if err := json.NewDecoder(r.Body).Decode(&newStudent); err != nil {
		http.Error(w, "잘못된 요청입니다.", http.StatusBadRequest)
		return
	}

	// 2
	query := `INSERT INTO students (name, email, borrowed_books)
			  VALUES ($1, $2, $3)
			  RETURNING id`

	err := studentsDB.QueryRowContext(r.Context(), query,
		newStudent.Name,
		newStudent.Email,
		pq.Array(newStudent.BorrowedBooks)).Scan(&newStudent.ID)

	if err != nil {
		log.Printf("DB 저장 실패 (%s): %v", "CreateStudent", err)
		http.Error(w, "DB 저장 실패", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(newStudent)
	if err != nil {
		http.Error(w, "JSON 변환 실패", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func GetStudentByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	query := "SELECT id, name, email, borrowed_books FROM students WHERE id=$1"

	var s Student
	err := studentsDB.QueryRowContext(r.Context(), query, id).Scan(
		&s.ID, &s.Name, &s.Email, pq.Array(&s.BorrowedBooks),
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "학생을 찾을 수 없음", http.StatusNotFound)
		} else {
			log.Printf("DB 조회 실패 (%s): %v", "GetStudentByID", err)
			http.Error(w, "DB 조회 실패", http.StatusInternalServerError)
		}
		return
	}

	jsonData, err := json.Marshal(s)
	if err != nil {
		http.Error(w, "JSON 변환 실패", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func UpdateStudent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updatedStudent Student
	if err := json.NewDecoder(r.Body).Decode(&updatedStudent); err != nil {
		http.Error(w, "잘못된 요청입니다.", http.StatusBadRequest)
		return
	}

	query := `UPDATE students SET name=$1, email=$2, borrowed_books=$3
			  WHERE id=$4`

	res, err := studentsDB.ExecContext(r.Context(), query,
		updatedStudent.Name,
		updatedStudent.Email,
		pq.Array(updatedStudent.BorrowedBooks),
		id,
	)
	if err != nil {
		log.Printf("DB 수정 실패 (%s): %v", "UpdateStudent", err)
		http.Error(w, "DB 수정 실패", http.StatusInternalServerError)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "DB 결과 확인 실패", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "학생을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	updatedStudent.ID = id
	jsonData, err := json.Marshal(updatedStudent)
	if err != nil {
		http.Error(w, "JSON 변환 실패", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func DeleteStudent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	query := "DELETE FROM students WHERE id =$1"

	res, err := studentsDB.ExecContext(r.Context(), query, id)
	if err != nil {
		log.Printf("DB 수정 실패 (%s): %v", "DeleteStudent", err)
		http.Error(w, "DB 삭제 실패", http.StatusInternalServerError)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "DB 결과 확인 실패", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "학생을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
