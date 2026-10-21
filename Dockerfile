# =========================================================================
# Multi-Stage Build Dockerfile for GCP Cloud Run
# 寃쎈웾 諛?蹂댁븞???꾪빐 CGO_ENABLED=0 ?뺤쟻 諛붿씠?덈━? 寃쎈웾 Alpine ?고??꾩쓣 ?ъ슜?⑸땲??
# =========================================================================

# Stage 1: Build binary
FROM golang:alpine AS builder

WORKDIR /app

# TLS ?몄쬆??諛?git ?ㅼ튂
RUN apk add --no-cache ca-certificates git

# ?섏〈???ㅼ슫濡쒕뱶 罹먯떆 ?덉씠??COPY go.mod go.sum ./
RUN go mod download

# ?뚯뒪 肄붾뱶 蹂듭궗
COPY . .

# ?뺤쟻 而댄뙆??(?붾쾭洹??щ낵 ?쒓굅濡??대?吏 ?⑸웾 理쒖냼??
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/blindtag-server ./cmd/server/main.go

# Stage 2: Minimal runtime image
FROM alpine:3.21

WORKDIR /app

# HTTPS ?듭떊 諛??쒓컙? ?ㅼ젙???꾪븳 ?⑦궎吏 ?ㅼ튂
RUN apk --no-cache add ca-certificates tzdata

# 鍮꾨（???ъ슜???앹꽦 (而⑦뀒?대꼫 蹂댁븞 沅뚯옣?ы빆)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# 鍮뚮뱶??諛붿씠?덈━ 蹂듭궗
COPY --from=builder /app/blindtag-server /app/blindtag-server

# Cloud Run 湲곕낯 ?ы듃
EXPOSE 8080
ENV PORT=8080
ENV APP_ENV=production

ENTRYPOINT ["/app/blindtag-server"]