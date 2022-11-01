package utils

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/security"
	"database/sql"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const X_FORWARDED_FOR = "X-Forwarded-For"

func UpdateLogs(claims security.Claims, req *http.Request, pool *DatabasePool, config *config.Config) {
	WriteLater(pool, func(Database *sql.DB) {
		stmt, err := Database.Prepare("insert into endpoint_audit_logs" +
			"(id,user_id,method,endpoint,body,ip_address)" +
			"VALUES ($1,$2,$3,$4,$5,$6);")

		if err != nil {
			log.Println(err)
			return
		}
		defer stmt.Close()

		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			log.Println(err)
			return
		}

		ipAddr := strings.Split(req.RemoteAddr, ":")[0] // remove port number

		if config.XForward {
			x_forwarded_for := req.Header.Get(X_FORWARDED_FOR)
			if x_forwarded_for != "" {
				ipAddr = x_forwarded_for
			}
		}

		if ipAddr == "" {
			ipAddr = "0.0.0.0"
		}
		_, err = stmt.Exec(uuid.New().String(), claims.Uuid, req.Method, req.RequestURI, body, ipAddr)

		if err != nil {
			log.Println(err)
		}
	})
}
