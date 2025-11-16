# MYGOAPI
간단한 REST API를 GO를 구축하여, API 설계에 대한 흐름을 확인합니다.

net/http 표준 패키지를 활용해서 간단하게 활용합니다.

1. 라우팅 설정: /users 경로에 대해 GET 요청이 오면 GetAllUsers 함수를 실행하도록 설정합니다.
2. 데이터 모델: User 구조체(struct)를 정의하여 데이터 모델을 만듭니다.
3. 핸들러 함수: 각 함수(GetAllUsers, CreateUser 등) 내에서 JSON 요청을 파싱하고, 비즈니스 로직을 처리하며, 마지막으로 JSON 응답과 적절한 HTTP 상태 코드를 클라이언트에게 전송합니다.

## 참고 문서

https://pkg.go.dev/net/http
