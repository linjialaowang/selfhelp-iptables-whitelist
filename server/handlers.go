package server

import (
	"fmt"
	"log"
	"net/http"
	"selfhelp-iptables-whitelist/config"
	"selfhelp-iptables-whitelist/ipt"
	"selfhelp-iptables-whitelist/utils"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
)

func checkKey(req *http.Request, privilege bool, apiName string) (result bool) {
	if err := req.ParseForm(); err == nil {
		key := req.Form["key"]
		remoteIP := strings.Split(req.RemoteAddr, ":")[0]
		now := time.Now().Format("2006-01-02 15:04:05")
		//golang没有三元运算符
		if len(key) > 0 {
			if privilege {
				if key[0] == config.GetConfig().AdminKey {
					utils.CmdColorGreen.Printf("%s IP: %s 尝试请求API: %s 已允许\n", now, remoteIP, apiName)
					result = true
				} else {
					result = false
					utils.CmdColorYellow.Printf("%s IP: %s 尝试请求API: %s 已拒绝 错误的KEY: %s\n", now, remoteIP, apiName, key[0])
				}
			} else {
				if key[0] == config.GetConfig().UserKey && key[0] == config.GetConfig().AdminKey {
					utils.CmdColorGreen.Printf("%s IP: %s 尝试请求API: %s 已允许\n", now, remoteIP, apiName)
					result = true
				} else {
					utils.CmdColorYellow.Printf("%s IP: %s 尝试请求API: %s 已拒绝 错误的KEY: %s\n", now, remoteIP, apiName, key[0])
					result = false
				}
			}
		} else {
			color.Set(color.FgRed)
			utils.CmdColorYellow.Printf("%s IP: %s 尝试请求API: %s 已拒绝 未设置KEY\n", now, remoteIP, apiName, key[0])
			color.Unset()
			result = false
		}
	} else {
		log.Println("KEY解析错误")
	}
	return
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "SelfHelp iptables Whitelist\n/api/add?key=yourkey\n/api/list?key=yourkey"+
		"\n/api/remove/ip?key=yourkey\n/api/log?key=yourkey\n/api/record?key=yourkey")
}

func AddWhitelist(w http.ResponseWriter, req *http.Request) {
	keyAuthentication := checkKey(req, false, "AddWhitelist")
	remoteIP := strings.Split(req.RemoteAddr, ":")[0]
	if keyAuthentication {
		ipt.AddIPWhitelist(remoteIP)
		utils.CmdColorGreen.Println("添加ip白名单 " + remoteIP)
		fmt.Fprintf(w, "添加ip白名单:"+remoteIP)
		ipt.WhiteIPs[remoteIP] = true
	} else {
		fmt.Fprintf(w, "key错误")
	}
}

func AddBlackList(w http.ResponseWriter, req *http.Request) {
	keyAuthentication := checkKey(req, true, "AddBlackList")
	if keyAuthentication {
		vars := mux.Vars(req)
		addIP := vars["ip"]
		ipt.AddIPBlacklist(addIP)
		utils.CmdColorGreen.Println("添加ip黑名单 " + addIP)
		fmt.Fprintf(w, "添加ip黑名单:"+addIP)
		ipt.BlackIPs[addIP] = true
	} else {
		fmt.Fprintf(w, "key错误")
	}
}

func RemoveWhitelist(w http.ResponseWriter, req *http.Request) {
	keyAuthentication := checkKey(req, true, "RemoveWhitelist")
	if keyAuthentication {
		vars := mux.Vars(req)
		removeIP := vars["ip"]
		utils.CmdColorGreen.Println("移除ip白名单 " + removeIP)
		fmt.Fprintf(w, "移除ip白名单:"+removeIP)
		ipt.DelIPWhitelist(removeIP)
		if _, exist := ipt.WhiteIPs[removeIP]; exist {
			delete(ipt.WhiteIPs, removeIP)
		}
	} else {
		fmt.Fprintf(w, "key错误")
	}
}

func RemoveBlacklist(w http.ResponseWriter, req *http.Request) {
	keyAuthentication := checkKey(req, true, "RemoveBlacklist")
	if keyAuthentication {
		vars := mux.Vars(req)
		removeIP := vars["ip"]
		utils.CmdColorGreen.Println("移除ip黑名单 " + removeIP)
		fmt.Fprintf(w, "移除ip黑名单:"+removeIP)
		ipt.DelIPBlacklist(removeIP)
		if _, exist := ipt.BlackIPs[removeIP]; exist {
			delete(ipt.BlackIPs, removeIP)
		}
	} else {
		fmt.Fprintf(w, "key错误")
	}
}

func ShowWhitelist(w http.ResponseWriter, req *http.Request) {
	keyAuthentication := checkKey(req, true, "ShowWhitelist")
	if keyAuthentication {
		//获取ips
		var ips string
		for ip, _ := range ipt.WhiteIPs {
			ips = fmt.Sprintln(ips, ip)
		}
		fmt.Fprintf(w, ips)
	} else {
		fmt.Fprintf(w, "key错误")
	}
}

func ShowBlacklist(w http.ResponseWriter, req *http.Request) {
	keyAuthentication := checkKey(req, true, "ShowBlacklist")
	if keyAuthentication {
		//获取ips
		var ips string
		for ip, _ := range ipt.BlackIPs {
			ips = fmt.Sprintln(ips, ip)
		}
		fmt.Fprintf(w, ips)
	} else {
		fmt.Fprintf(w, "key错误")
	}
}

func GetLogs(w http.ResponseWriter, req *http.Request) {
	keyAuthentication := checkKey(req, true, "GetLogs")
	if keyAuthentication {
		//获取日志
		ipLogs := utils.ExecCommand(`cat ` + ipt.KernLogURL + `| grep netfilter`)
		fmt.Fprintf(w, ipLogs)
	} else {
		fmt.Fprintf(w, "key错误")
	}
}

func Reset(w http.ResponseWriter, req *http.Request) {
	keyAuthentication := checkKey(req, true, "Reset")
	if keyAuthentication {
		//获取日志
		ipt.ResetIPWhitelist()
		fmt.Fprintf(w, "已进行重置")
	} else {
		fmt.Fprintf(w, "key错误")
	}
}

func Vnstat(w http.ResponseWriter, req *http.Request) {
	keyAuthentication := checkKey(req, true, "Vnstat")
	param := strings.TrimSpace(req.URL.Query().Get("param"))
	if keyAuthentication {
		//获取日志
		// 需要检查参数的合法性
		var validParams = []string{"-5", "-h", "--hours", "--hoursgraph", "-hg", "-d", "--days", "-m", "--months", "-y", "--years", "-t", "--top", ""}
		var valid bool
		for _, p := range validParams {
			if param == p {
				valid = true
				break
			}
		}
		if valid {
			stat := utils.ExecCommand("vnstat " + param)
			fmt.Fprintf(w, stat)
		} else {
			fmt.Fprintf(w, "无效参数: "+param)
		}

	} else {
		fmt.Fprintf(w, "key错误")
	}
}

//只输出ip和探测数量

func GetRecords(w http.ResponseWriter, req *http.Request) {
	keyAuthentication := checkKey(req, true, "GetRecords")
	var whitelistStrBuilder, nowhitelistStrBuilder strings.Builder
	if keyAuthentication {
		for ip, count := range ipt.RecordedIPs {
			if _, exist := ipt.WhiteIPs[ip]; exist {
				whitelistStrBuilder.WriteString(ip)
				whitelistStrBuilder.WriteString(" 记录次数: ")
				whitelistStrBuilder.WriteString(strconv.Itoa(count))
				whitelistStrBuilder.WriteString(" [白名单]\n")
			} else {
				nowhitelistStrBuilder.WriteString(ip)
				nowhitelistStrBuilder.WriteString(" 记录次数: ")
				nowhitelistStrBuilder.WriteString(strconv.Itoa(count))
				nowhitelistStrBuilder.WriteString("\n")
			}
		}

		strBuilder := strings.Builder{}
		strBuilder.WriteString(fmt.Sprintf("共有个%d个ip被记录,其中%d个ip添加了白名单,%d个ip没有添加白名单\n", len(ipt.RecordedIPs), len(ipt.WhiteIPs), len(ipt.RecordedIPs)-len(ipt.WhiteIPs)))
		strBuilder.WriteString(whitelistStrBuilder.String())
		strBuilder.WriteString(nowhitelistStrBuilder.String())
		fmt.Fprintln(w, strBuilder.String())
	} else {
		fmt.Fprintf(w, "key错误")
	}
}
