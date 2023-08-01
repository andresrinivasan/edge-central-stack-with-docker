// Copyright (C) 2021 IOTech Ltd

package xpert

import (
	"fmt"
	"strings"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	encryptedKey = `
-----BEGIN RSA PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: DES-EDE3-CBC,8B670A7E72E9D7FE

veBUwC9MTOgEdUUl2ef9PiZ/xZypiYbvGD9Mnc1vKUrQHXZ2TAevTGs4PDAE7DMs
y4hD/eEFWIGHdfS3Uk+YP3usrjYrdBP7eUrxUeig7V5w+W1G5IFRDTKbEerfjuM7
WjhU5+zds56iuTDCDMg7nxjocMUJoF7wnA8ik8KejBa3jkYz07arMnl0uxacKX9E
JSJhzCIB442wVEF716JV2aVNso7F4G5EJzex+CP3W4pqkGjzBPOOpvHilWQ8sHf9
jvhLIX6McaW/dGmuI2n3I9OEzJqJqIoHni5k9pyJohsCppXGfat5dPl7S7R5zmnw
qk4P9FEAhBmj/qxy4v8e71qm7lC16UOizBJneWeZhx6esWNWx0sAnlmFQ4VQNSjA
/MN5JMHfY5kpItCOFtKsm3JdDP9nXEgVxufO7a5yHD+Ur8hyVJ3qyOHIh/h1kvt/
VX4/gJIadyBjkNJz/NSvcWst9TxajNAY2GHa0W/geHHTndUMSmKR+sOYH0E/uhTW
7g7BzFcliV2QZgCf2SDA8ClReJsaqljfl0Iama10A+ebh5jTbTDokMhQwVEA0YGi
/lcmdC+m//PGpbzaEOZ29EgBtYmbOMdgr81sx2KpWTwHa7ZsQMAsMStXRBuUj+Xv
PldXqk1QdVljpYwo6UmZ6zPUf+AHRc4t9GR1rutzqPysJ5vzBdpCy5XmpLKlAgWJ
dtd34spavwDchVkRFj54expQDfUEwW+jXnQEJy7qt2HZwFvbGdjtRUEaR0FMV8ea
FB2K9/NPPhiPO1bcf/YimbJiiFAwMPo7LEebb8mIUjogRQGCgPRMs8iLk+WX/QJQ
9iV2hSovOZ64ASfT92sxonNSAEMtMmyBCkBuvqUArtV94Qr46Rho/QxEGkDjnLv6
2xWmmEGnTAsfsMi5kATWE16h/ctHnBi3d4vN298Nm0Bcm4YGbBUOxU6s96YgCJWv
bAxKrCzuS565oEw2qNcOK0VzNy9rF792wULsREg1cTIcOLbHAwKpaXojO+jYhxqJ
P6DgTkaSaoR7qCAPu1LA/t7cK1pMzUisbqgpLwZ7/RTsWJeweVw8jsFv2sdz6Kl3
kF3tfsQ/PMtwulMCB8f7efa2EpR63zo4L20xmmtcKPC2IO6tQdbvGndYS/oob81s
Pvq0tJTevTutZ4ENpkVXcGS7hdP/C4chQRIe4ai6Jk2w2Ul2/khvtgNaPW+VQwcn
aT48EUzTbKP/ZUw+iCTB0UvSHLpMaqDdRmR8SxfBn+LUTJQeHlzFhL/tOZjzoFyh
wjxlZrHjo1P7XKr3p7bk+aY2rDodL3kjIw1dYLlDW1wR2yTP5Oon8zSElBLsuxGf
zUshfLqv++jRqgbDaYJwPbXo21RpPaRj4IkAy9lcGvxcRsAVi0hlaGtACgoIuiZp
6k8phVLaQPm09bNdwfLxcGWOOtGLdQ6sxwspjTlbX8F0bHZGHw79WeBZygSAOENO
kYlRE9QT+5I+TgmpRAGH3pj36gi6ESlvG2yFEz2p2Z2l72j5SgoKhEdJsA4jkz+R
B+emis0H4fT3BnMzIYvDkyOKL5G/Pj+VTpWBl0MylT93+RGYCWECgg==
-----END RSA PRIVATE KEY-----
`
	decryptedClientKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAnzzp4FtdLfaQ/B799nJiW5pcQ6YGLDrcFkKPDPVhPTEy54Ch
mkviLmEHHbzsj94IH+G5JqZ3qnaW4l00G7CmhJOpL2vQRpCjLdG/ZNk0S5uLJ94Z
aeA9UWm76A3CwZhucfpwdZ+ptLoR89AxikUBuF5Md0y/LFAg25sinz6dL2KJAQh+
gpAT0sslj0Yz69ma04swnD1LEz0y7BVpvZz1xpRML3KAmA5GjvFCw1SqUJlFxrsS
U+Hl3WcEiPHnaU5rb2fHYXua1hZ6Lm7UJo2bVESALPaWG6sYYffz9L7PFnCQqhGR
uQv/in2Zy9ROjKKo3qqzH/BnIXA6r2BAVdTYtwIDAQABAoIBAD3GsvEQYODhBDRb
jakbjR7+jobMFR75osKcBcVAOP41ZQs88vTaNaBKkikuTxQtTjeYKW1eLZSbN0QQ
ZpPLf351jrBQAlgt6rBu6/Ki9U/TwzOvTWquzPsVqwmGtSTIDyj2wMRRMdRkT2yo
O1/qD5XIN6AczRnS6DxqPg9Lik2ELutoTD9l4uMi6USE25Age3DeLeJ1c1SrEM0r
R8Ywh/TPSVeBxukVk/T5OKOD+hOp8Bthrl2yhY2YOur4bNv+6Sx5wJ5606w5QQjZ
LtLD0VPuxlIe7yRlfcCXAEqc37yJycZQkIdMCN8DbxNffszWiJbzMx8JFvv7O/dk
1sGyjrECgYEAz8myvseCM+/WTPKOfTVwG15PUKUKuZN5QLt7WGpMEB9x9e5f+ZsH
Zauy/PQ1OyHORgOgMqY1dd4gBFG9cxpR5tNq8Jxd1io2dK5oXq8VBy5zWyTunECp
cYUVxeQBRIkZA0/q4fOD1U5B69+VTSOKWnpW0y1VJNWAfLsCbnpu6V8CgYEAxC9r
LcvOsJU8tqAjyBoAoBPSYEvnObUvx0tVcGkm+BC5N4GNTWqM+YfWLMpqGUoKAIYu
PRfcT+/VfWEWeJf3tBHk9vznNrsHh2rMRj6QyxX1IMsfSvJPNePF6GnB23N1txiz
R315QuxvQEPhJGLjocHHUC68Q6929PRJs2Ii16kCgYEAuk61UE3+tqbTVYcer7Gc
ZU24fCyfYymRzLLNs8cLkGFBgytLLrkMduLux9QHbo+vLiPOHdvdj2Os/XJ1FaGB
0h+6gScTFBYhYZmHx23gwuGpWQ3STJPF2h1kGl2HrXXn0Yp0pkf76uQSQ3XjnpjB
UsLi2tKIx1APtsbPNVPd4q0CgYBe2J30cgfnDv9fO2SRJSEQQwT+UTPkjlge/ai8
w9l3LH6e+x8ZQl4NdUJyPRm2SDk1r6lDF/oHG2gXSYzXmIDEqbIMRpBxwVIOge9o
Nm9B/8eWpxzl2ue4ofnYNujl85gBgQuLkHnDhRLz+t0p/jUWytxVQ4L5JidYnZHU
C6nUoQKBgHpwoPhUR9sWwrpPdeFj2Bb03mLQn2vRJ64IPmo/hJE3oQ7lAa2veDIB
Nf7qrclUuGwpB7mHAxTC+7D6UUftuV7vvnSXkyIgPbx80yaTGf/BZZQpl06a/6in
R9FQ9hSMc5BIuG30eXLvvgZsUOY1ETDROxh6RnmNL1Q9rn7l3poL
-----END RSA PRIVATE KEY-----
`
	kafkaDecryptedPassword = "abcdefgh"
	kafkaClientKey         = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDpL/BLBImzFTnq
XMZD+q3C5j9XWBdIue4YKIgPfydECxu+VEP8TIk53Z+0iULCAVOR+xJvt9ZGyF/P
Y/A8mx0Sx4XdgOOzlm+YXLE3NCA+dzLYV7EXNhAKFD3sTr2lIOBpaCT2JjuU/gb/
EwNZQShJ31SJNY3k0negVaqB85vZQ589UE00DpfF1qJVUbn85PyukUkG6VpYXSQ8
2C3Swx1JSmMbPJzn/NqxjWa6OhIt6h7LDvlcvunmDNiexIt/bA6TG4AtidTwmSw4
idi5tssZNwld+g2GTaVxVJFOx8fDizs/xfDuhQSQzV/Y8kJ5UVKDxCEZbYgV1Nhg
MZCxJ/jjAgMBAAECggEASXxHAI0ci+gbiUTVYmTkT3BZ87+aDtwxMUHMpv2ONT4+
7vsFNcQ01pyGENHUzOi4GmACDlzj1QieUPAQrDjBr8Ja6FQO7fBxmJVVb60ooCbW
SiFQeJ0b7uE0Jn0l/Jzgu7cLNtsTmb94Gvg17PHArY8Ix/itj7fX5Ro6EpvfuFai
GU76/ZYGhbK6D1J5bV9U5NFmUVHNwSyhZRH1b/ozna7ciJCiTiiJyazyPWiibGDg
d3KZ18Zjs/ULeAuSN68EZ+KwtcgR/7awvfucmqaFinUn90xKRXXWw+xxBWfGcc9V
wd5murFvmSf7fec2n8vz6WrQVADx0ny4Th2qCSI+SQKBgQD5Cu8T/XfxLQ9TOgFc
p8b29UXlhvtuC+EJmg3tV92JJeG/8N35epoKvt7VufKZm6cUCLwiglzBJPvINp2h
mrop+g1MNXnDpgraUbzufOGDD6+zL6gXhZIzHTWWhkO8KwBFrM1f/7Be8V8dTyGb
ZC9+GblG3nkdKghhXg80m3AtrQKBgQDvs5y0bYYzbFNw2JiIeFKpcGG5KsWv3zpX
B4wImlXycXGPLdNtrxIR7Ovf5IpCR6BIBmXQF8yao0itLnBjwr3EPCbM3xnzBkyF
Iqcm0pRA2SQk7cZSJw6dwJPH+hMgQ5Nn5LQcpsmDKz4PGWsN5sETYLMiVgKjIN5d
wttLrQxyzwKBgHF86w/3/LV57DboAwDfMAsQIIcFKQSwAx/mBRy4YqsCCUr3j6AF
n7bv3goVT5lyVgQKKvmq4Gvf16EYSmL/aICCg5bL864Vt3JftzIS1I1uE4obWIVH
iCUk1Wu/yZQxIFGf+oMZuJy7b7WiftUaJY5YWJcUAKsqoWEFhPZbMxaNAoGAWDy5
Gd4bgcCFssu40rvgSglZn+0z2nsFIdZgYSZXLyk9kWRgKUdCEqExbzjVAHMXeIwK
XKD2K5KiBUZMDx03+A3ghpg2GDUgY/4OpAbuljSYzpNM5x8DjWS/weS3t6/Iin0x
JD7tfUCk1rAXrYVdW8HED4az79MAqGk7is8H/xcCgYEA58sJZEqVWT6jAGVZEZLL
eqJkgQAwUjfRNl1ciqdeDtDhbqzuxImFVuJSg4ObaUCBFQM9bgLGrNR+rht7QZSN
IpNwn9y9hGwVZEeaDqj2+NHpvPR7FbEl5ZLY/38QwSm1YM7++7qMYCC/kzqQ4kTs
xjptD/Pt0aGvRDIJp62GCbE=
-----END PRIVATE KEY-----
`
	kafkaClientCert = `
-----BEGIN CERTIFICATE-----
MIIDgzCCAmugAwIBAgIUMZbvj2SvMy3A+xLY+Njth8fRr/cwDQYJKoZIhvcNAQEF
BQAwUTELMAkGA1UEBhMCVFcxEzARBgNVBAgMClNvbWUtU3RhdGUxDzANBgNVBAcM
BlRhaXBlaTEPMA0GA1UECgwGSU9UZWNoMQswCQYDVQQLDAJSRDAeFw0yMDA4MDIx
MDE3NTlaFw0yMTA4MDIxMDE3NTlaMFExCzAJBgNVBAYTAlRXMRMwEQYDVQQIDApT
b21lLVN0YXRlMQ8wDQYDVQQHDAZUYWlwZWkxDzANBgNVBAoMBklPVGVjaDELMAkG
A1UECwwCUkQwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDpL/BLBImz
FTnqXMZD+q3C5j9XWBdIue4YKIgPfydECxu+VEP8TIk53Z+0iULCAVOR+xJvt9ZG
yF/PY/A8mx0Sx4XdgOOzlm+YXLE3NCA+dzLYV7EXNhAKFD3sTr2lIOBpaCT2JjuU
/gb/EwNZQShJ31SJNY3k0negVaqB85vZQ589UE00DpfF1qJVUbn85PyukUkG6VpY
XSQ82C3Swx1JSmMbPJzn/NqxjWa6OhIt6h7LDvlcvunmDNiexIt/bA6TG4AtidTw
mSw4idi5tssZNwld+g2GTaVxVJFOx8fDizs/xfDuhQSQzV/Y8kJ5UVKDxCEZbYgV
1NhgMZCxJ/jjAgMBAAGjUzBRMB0GA1UdDgQWBBR9XM4em5Y/i+K9pAxs2EtNy4G/
KjAfBgNVHSMEGDAWgBR9XM4em5Y/i+K9pAxs2EtNy4G/KjAPBgNVHRMBAf8EBTAD
AQH/MA0GCSqGSIb3DQEBBQUAA4IBAQAb3FuP1t2lM09E7GSZJSwYDfLO16q6/oLa
ulnLu8MoQX1L9ME8s7UhrpCge4iHLvzmWMx4Cp07Ap2SdyLnoxiwUJBj54mPIt7w
4gxNBP5+4KjeQmWkNTUFCfiLH9ZNsqivy/T48ZzXPE2HUPd3mwuOj1Ev+FXZUh0+
ZLzo7RZakqWO5H4WO4L1k8o6z+yV6ijjPzVRjXlhCi2G+A2TlnP5sgoTq9L0foQo
5Uz4ipNjXe7qrzW8p51s8WcPf7dulZZj73mfxYoHyHLdKdShDQuThZxl3Cnv5M96
TV28UN4M66kv4oY/FhPf50rGZp9wGIppzhPCqje7pFX1EXHFjTnE
-----END CERTIFICATE-----
`
	kafkaCACert = `-----BEGIN CERTIFICATE-----
MIIElDCCA3ygAwIBAgIQAf2j627KdciIQ4tyS8+8kTANBgkqhkiG9w0BAQsFADBh
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD
QTAeFw0xMzAzMDgxMjAwMDBaFw0yMzAzMDgxMjAwMDBaME0xCzAJBgNVBAYTAlVT
MRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxJzAlBgNVBAMTHkRpZ2lDZXJ0IFNIQTIg
U2VjdXJlIFNlcnZlciBDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB
ANyuWJBNwcQwFZA1W248ghX1LFy949v/cUP6ZCWA1O4Yok3wZtAKc24RmDYXZK83
nf36QYSvx6+M/hpzTc8zl5CilodTgyu5pnVILR1WN3vaMTIa16yrBvSqXUu3R0bd
KpPDkC55gIDvEwRqFDu1m5K+wgdlTvza/P96rtxcflUxDOg5B6TXvi/TC2rSsd9f
/ld0Uzs1gN2ujkSYs58O09rg1/RrKatEp0tYhG2SS4HD2nOLEpdIkARFdRrdNzGX
kujNVA075ME/OV4uuPNcfhCOhkEAjUVmR7ChZc6gqikJTvOX6+guqw9ypzAO+sf0
/RR3w6RbKFfCs/mC/bdFWJsCAwEAAaOCAVowggFWMBIGA1UdEwEB/wQIMAYBAf8C
AQAwDgYDVR0PAQH/BAQDAgGGMDQGCCsGAQUFBwEBBCgwJjAkBggrBgEFBQcwAYYY
aHR0cDovL29jc3AuZGlnaWNlcnQuY29tMHsGA1UdHwR0MHIwN6A1oDOGMWh0dHA6
Ly9jcmwzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydEdsb2JhbFJvb3RDQS5jcmwwN6A1
oDOGMWh0dHA6Ly9jcmw0LmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydEdsb2JhbFJvb3RD
QS5jcmwwPQYDVR0gBDYwNDAyBgRVHSAAMCowKAYIKwYBBQUHAgEWHGh0dHBzOi8v
d3d3LmRpZ2ljZXJ0LmNvbS9DUFMwHQYDVR0OBBYEFA+AYRyCMWHVLyjnjUY4tCzh
xtniMB8GA1UdIwQYMBaAFAPeUDVW0Uy7ZvCj4hsbw5eyPdFVMA0GCSqGSIb3DQEB
CwUAA4IBAQAjPt9L0jFCpbZ+QlwaRMxp0Wi0XUvgBCFsS+JtzLHgl4+mUwnNqipl
5TlPHoOlblyYoiQm5vuh7ZPHLgLGTUq/sELfeNqzqPlt/yGFUzZgTHbO7Djc1lGA
8MXW5dRNJ2Srm8c+cftIl7gzbckTB+6WohsYFfZcTEDts8Ls/3HB40f/1LkAtDdC
2iDJ6m6K7hQGrn2iWZiIqBtvLfTyyRRfJs8sjX7tN8Cp1Tm5gr8ZDOo0rwAhaPit
c+LJMto4JQtV05od8GiG7S5BNO98pVAdvzr508EIDObtHopYJeS4d60tbvVS3bR0
j6tJLp07kzQoH3jOlOrHvdPJbRzeXDLz
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDrzCCApegAwIBAgIQCDvgVpBCRrGhdWrJWZHHSjANBgkqhkiG9w0BAQUFADBh
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD
QTAeFw0wNjExMTAwMDAwMDBaFw0zMTExMTAwMDAwMDBaMGExCzAJBgNVBAYTAlVT
MRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5j
b20xIDAeBgNVBAMTF0RpZ2lDZXJ0IEdsb2JhbCBSb290IENBMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4jvhEXLeqKTTo1eqUKKPC3eQyaKl7hLOllsB
CSDMAZOnTjC3U/dDxGkAV53ijSLdhwZAAIEJzs4bg7/fzTtxRuLWZscFs3YnFo97
nh6Vfe63SKMI2tavegw5BmV/Sl0fvBf4q77uKNd0f3p4mVmFaG5cIzJLv07A6Fpt
43C/dxC//AH2hdmoRBBYMql1GNXRor5H4idq9Joz+EkIYIvUX7Q6hL+hqkpMfT7P
T19sdl6gSzeRntwi5m3OFBqOasv+zbMUZBfHWymeMr/y7vrTC0LUq7dBMtoM1O/4
gdW7jVg/tRvoSSiicNoxBN33shbyTApOB6jtSj1etX+jkMOvJwIDAQABo2MwYTAO
BgNVHQ8BAf8EBAMCAYYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUA95QNVbR
TLtm8KPiGxvDl7I90VUwHwYDVR0jBBgwFoAUA95QNVbRTLtm8KPiGxvDl7I90VUw
DQYJKoZIhvcNAQEFBQADggEBAMucN6pIExIK+t1EnE9SsPTfrgT1eXkIoyQY/Esr
hMAtudXH/vTBH1jLuG2cenTnmCmrEbXjcKChzUyImZOMkXDiqw8cvpOp/2PV5Adg
06O/nVsJ8dWO41P0jmP6P6fbtGbfYmbW0W5BjfIttep3Sp+dWOIrWcBAI+0tKIJF
PnlUkiaY4IBIqDfv8NZ5YBberOgOzW6sRBc4L0na4UU+Krk2U886UAb3LujEV0ls
YSEY1QSteDwsOoBrp+uvFRTp2InBuThs4pFsiv9kuXclVzDAGySj4dzp30d8tbQk
CAUw7C29C79Fv1C5qfPrmAESrciIxpg0X40KPMbp1ZWVbd4=
-----END CERTIFICATE-----
`
)

func TestNewKafkaSender(t *testing.T) {
	endpoint := KafkaEndpoint{
		ClientID:  "edgex",
		Address:   "localhost",
		Port:      9092,
		Topic:     "test",
		Partition: int32(0),
	}
	secretsConfig := KafkaSecretsConfig{
		AuthMode: messaging.AuthModeNone,
	}
	sender := NewKafkaSender(endpoint, secretsConfig, false)
	assert.NotNil(t, sender, "sender should not be nil")
}

func TestKafkaSendInvalidEndpoint(t *testing.T) {
	endpoint := KafkaEndpoint{
		ClientID:  "edgex",
		Address:   "abc",
		Port:      9091,
		Topic:     "test",
		Partition: int32(0),
	}
	secretsConfig := KafkaSecretsConfig{
		AuthMode: messaging.AuthModeNone,
	}
	sender := NewKafkaSender(endpoint, secretsConfig, false)
	continuePipeline, result := sender.KafkaSend(ctx, "fake data")

	assert.False(t, continuePipeline, "Pipeline should stop")
	assert.Error(t, result.(error), "Result should be an error")
	assert.True(t, strings.Contains(result.(error).Error(), "could not create Kafka producer"), "Shall fail to create Kafka producer")
}

func TestKafkaSendNoDataPassed(t *testing.T) {

	endpoint := KafkaEndpoint{
		ClientID:  "edgex",
		Address:   "localhost",
		Port:      9092,
		Topic:     "test",
		Partition: int32(0),
	}
	secretsConfig := KafkaSecretsConfig{
		AuthMode: messaging.AuthModeNone,
	}
	sender := NewKafkaSender(endpoint, secretsConfig, false)
	continuePipeline, result := sender.KafkaSend(ctx, nil)

	assert.False(t, continuePipeline, "Pipeline should stop")
	assert.Error(t, result.(error), "Result should be an error")
	assert.Equal(t, "no data received", result.(error).Error())
}

func TestKafkaSendPersistData(t *testing.T) {
	endpoint := KafkaEndpoint{
		ClientID:  "edgex",
		Address:   "localhost",
		Port:      9092,
		Topic:     "test",
		Partition: int32(0),
	}
	secretsConfig := KafkaSecretsConfig{
		AuthMode: messaging.AuthModeNone,
	}
	sender := NewKafkaSender(endpoint, secretsConfig, true)
	ctx.SetRetryData(nil)
	sender.setRetryData(ctx, []byte("fake data"))
	assert.NotNil(t, ctx.RetryData())
}

func TestKafkaSendGetSecrets(t *testing.T) {
	notFoundSecretPath := "notfound"
	returnSecretPath := "kafka"
	endpoint := KafkaEndpoint{
		ClientID:  "edgex",
		Address:   "localhost",
		Port:      9092,
		Topic:     "test",
		Partition: int32(0),
	}
	sender := NewKafkaSender(endpoint, KafkaSecretsConfig{}, false)
	tests := []struct {
		Name            string
		SecretPath      string
		ExpectedSecrets *kafkaSecrets
		ExpectingError  bool
	}{
		{"No Secrets found", notFoundSecretPath, nil, true},
		{"With Secrets", returnSecretPath, &kafkaSecrets{
			keyPEMBlock:       []byte(kafkaClientKey),
			certPEMBlock:      []byte(kafkaClientCert),
			caCertPEMBlock:    []byte(kafkaCACert),
			decryptedPassword: []byte(kafkaDecryptedPassword),
		}, false},
	}
	// setup mock secret client
	secrets := map[string]string{
		messaging.SecretClientKey:  kafkaClientKey,
		messaging.SecretClientCert: kafkaClientCert,
		messaging.SecretCACert:     kafkaCACert,
		KafkaDecryptPassword:       kafkaDecryptedPassword,
	}
	mockSecretProvider := &mocks.SecretProvider{}
	mockSecretProvider.On("GetSecret", notFoundSecretPath).Return(nil, fmt.Errorf("path (%v) doesn't exist in secret store", notFoundSecretPath))
	mockSecretProvider.On("GetSecret", returnSecretPath).Return(secrets, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSecretProvider
		},
	})
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			sender.secretsConfig = KafkaSecretsConfig{
				AuthMode:   messaging.AuthModeCert,
				SecretPath: test.SecretPath,
			}
			kafkaSecrets, err := sender.getSecrets(ctx)
			if test.ExpectingError {
				assert.Error(t, err, "Expecting error")
				return
			}
			require.Equal(t, test.ExpectedSecrets, kafkaSecrets)
		})
	}
}

func TestKafkaSendDecryptClientKey(t *testing.T) {
	endpoint := KafkaEndpoint{
		ClientID:  "edgex",
		Address:   "localhost",
		Port:      9092,
		Topic:     "test",
		Partition: int32(0),
	}
	sender := NewKafkaSender(endpoint, KafkaSecretsConfig{}, false)
	tests := []struct {
		Name            string
		Secrets         *kafkaSecrets
		ExpectedSecrets *kafkaSecrets
		ExpectingError  bool
	}{
		{"Decrypt with error", &kafkaSecrets{
			keyPEMBlock:       []byte("abc123"),
			certPEMBlock:      []byte(kafkaClientCert),
			caCertPEMBlock:    []byte(kafkaCACert),
			decryptedPassword: []byte(kafkaDecryptedPassword),
		}, nil, true},
		{"Decrypt with incorrect password", &kafkaSecrets{
			keyPEMBlock:       []byte(encryptedKey),
			certPEMBlock:      []byte(kafkaClientCert),
			caCertPEMBlock:    []byte(kafkaCACert),
			decryptedPassword: []byte("incorrect password"),
		}, nil, true},
		{"Decrypt with correct password", &kafkaSecrets{
			keyPEMBlock:       []byte(encryptedKey),
			certPEMBlock:      []byte(kafkaClientCert),
			caCertPEMBlock:    []byte(kafkaCACert),
			decryptedPassword: []byte(kafkaDecryptedPassword),
		}, &kafkaSecrets{
			keyPEMBlock:       []byte(encryptedKey),
			certPEMBlock:      []byte(kafkaClientCert),
			caCertPEMBlock:    []byte(kafkaCACert),
			decryptedPassword: []byte(kafkaDecryptedPassword),
		}, false},
		{"no need for decryption", &kafkaSecrets{
			keyPEMBlock:       []byte(decryptedClientKey),
			certPEMBlock:      []byte(kafkaClientCert),
			caCertPEMBlock:    []byte(kafkaCACert),
			decryptedPassword: []byte(kafkaDecryptedPassword),
		}, &kafkaSecrets{
			keyPEMBlock:       []byte(decryptedClientKey),
			certPEMBlock:      []byte(kafkaClientCert),
			caCertPEMBlock:    []byte(kafkaCACert),
			decryptedPassword: []byte(kafkaDecryptedPassword),
		}, false},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := sender.decryptClientKey(test.Secrets)
			if test.ExpectingError {
				assert.Error(t, err, "Expecting error")
				return
			}
			assert.NoError(t, err, "expect no errors")
		})
	}
}

func TestKafkaSendValidateSecret(t *testing.T) {
	endpoint := KafkaEndpoint{
		ClientID:  "edgex",
		Address:   "localhost",
		Port:      9092,
		Topic:     "test",
		Partition: int32(0),
	}
	sender := NewKafkaSender(endpoint, KafkaSecretsConfig{}, false)
	tests := []struct {
		Name           string
		secretConfig   KafkaSecretsConfig
		Secrets        kafkaSecrets
		ExpectingError bool
	}{
		{"Validate secret with authmode none", KafkaSecretsConfig{
			AuthMode: messaging.AuthModeNone,
		}, kafkaSecrets{}, false},
		{"Validate secret with authmode clientcert but no clientkey", KafkaSecretsConfig{
			AuthMode:   messaging.AuthModeCert,
			SecretPath: "kafka",
		}, kafkaSecrets{
			certPEMBlock:      []byte(kafkaClientCert),
			caCertPEMBlock:    []byte(kafkaCACert),
			decryptedPassword: []byte(kafkaDecryptedPassword),
		}, true},
		{"Validate secret with authmode clientcert but no clientcert", KafkaSecretsConfig{
			AuthMode:   messaging.AuthModeCert,
			SecretPath: "kafka",
		}, kafkaSecrets{
			keyPEMBlock:       []byte(kafkaClientKey),
			caCertPEMBlock:    []byte(kafkaCACert),
			decryptedPassword: []byte(kafkaDecryptedPassword),
		}, true},
		{"Validate secret with authmode clientcert and sufficient clientkey/clientcert", KafkaSecretsConfig{
			AuthMode:   messaging.AuthModeCert,
			SecretPath: "kafka",
		}, kafkaSecrets{
			keyPEMBlock:       []byte(kafkaClientKey),
			certPEMBlock:      []byte(kafkaClientCert),
			caCertPEMBlock:    []byte(kafkaCACert),
			decryptedPassword: []byte(kafkaDecryptedPassword),
		}, false},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			sender.secretsConfig = test.secretConfig
			err := sender.validateSecrets(test.Secrets)
			if test.ExpectingError {
				assert.Error(t, err, "Expecting error")
				return
			}
			assert.NoError(t, err, "expect no errors")
		})
	}
}
