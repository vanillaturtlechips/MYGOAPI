// main.js

// 버튼 클릭 시 API 호출
document.getElementById('loadButton').addEventListener('click', () => {
    
    // [중요] Nginx의 /api/ 경로를 통해 Go API를 호출
    fetch('/api/students') 
        .then(response => response.json())
        .then(students => {
            const listElement = document.getElementById('studentList');
            listElement.innerHTML = ''; // 목록 초기화

            if (students && students.length > 0) {
                students.forEach(student => {
                    const item = document.createElement('li');
                    item.textContent = `[ID: ${student.id}] ${student.name} (${student.email})`;
                    listElement.appendChild(item);
                });
            } else {
                const item = document.createElement('li');
                item.textContent = '학생이 없습니다.';
                listElement.appendChild(item);
            }
        })
        .catch(error => {
            console.error('Error fetching students:', error);
            const listElement = document.getElementById('studentList');
            listElement.innerHTML = '<li>데이터 로드 실패!</li>';
        });
});