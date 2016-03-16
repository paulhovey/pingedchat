package mailWebhooks

type MandrillWebHook []MandrillWebHookContent

type MandrillWebHookContent struct {
	Ts    uint64                 `json:"ts"`
	Event string                 `json:"event"`
	Msg   MandrillWebHookMessage `json:"msg"`
}

type MandrillWebHookMessage struct {
	RawMsg      string                      `json:"raw_msg"`
	Headers     MandrillWebhookHeaders      `json:"headers"`
	Text        string                      `json:"text"`
	Html        string                      `json:"html"`
	FromEmail   string                      `json:"from_email"`
	FromName    string                      `json:"from_name"`
	To          [][]string                  `json:"to"`
	Email       string                      `json:"email"`
	Subject     string                      `json:"subject"`
	Tags        []string                    `json:"tags"`
	Sender      string                      `json:"sender"`
	TextFlowed  bool                        `json:"text_flowed"`
	Attachments map[string]MandrillWebhookAttachment `json:"attachments"`
	Images      []MandrillWebhookAttachment `json:"images"`
	SpamReport  MandrillWebhookSpamReport   `json:"spam_report"`
	Dkim        MandrillWebhookDkim         `json:"dkim"`
	Spf         MandrillWebhookSpf          `json:"spf"`
}

type MandrillWebhookHeaders struct {
	ContentType        string   `json:"Content-Type"`
	Date               string   `json:"Date"`
	DkimSignature      string `json:"Dkim-Signature"`
	DomainKeySignature string   `json:"Domainkey-Signature"`
	From               string   `json:"From"`
	ListUnsubscribe    string   `json:"List-Unsubscribe"`
	MessageId          string   `json:"Message-Id"`
	MimeVersion        string   `json:"Mime-Version"`
	Received           []string `json:"Received"`
	Sender             string   `json:"Sender"`
	Subject            string   `json:"Subject"`
	To                 string   `json:"To"`
	XReportAbuse       string   `json:"X-Report-Abuse"`
}

type MandrillWebhookAttachment struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	IsBase64  bool   `json:"base64"`
}

type MandrillWebhookSpamReport struct {
	Score        float32 `json:"score"`
	MatchedRules []struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Score       float32 `json:"score"`
		Spf         struct {
			Result string `json:"result"`
			Detail string `json:"detail"`
		} `json:"spf,omitempty"`
	} `json:"matched_rules"`
}

type MandrillWebhookDkim struct {
	Signed bool `json:"signed"`
	Valid  bool `json:"valid"`
}

type MandrillWebhookSpf struct {
	Detail string `json:"detail"`
	Result string `json:"result"`
}

// for database storage and retrieval
// type MailAttachment struct {
	
// }

// testing playground:  https://play.golang.org/p/np6luvh6D1
// EXAMPLE POST
// [{
//     "event": "inbound",
//     "msg": {
//         "dkim": {
//             "signed": true,
//             "valid": true
//         },
//         "email": "example@pinged.email",
//         "from_email": "example.sender@mandrillapp.com",
//         "headers": {
//             "Content-Type": "multipart\/alternative; boundary=\"_av-7r7zDhHxVEAo2yMWasfuFw\"",
//             "Date": "Fri, 10 May 2013 19:28:20 +0000",
//             "Dkim-Signature": ["v=1; a=rsa-sha1; c=relaxed\/relaxed; s=mandrill; d=mail115.us4.mandrillapp.com; h=From:Sender:Subject:List-Unsubscribe:To:Message-Id:Date:MIME-Version:Content-Type; i=example.sender@mail115.us4.mandrillapp.com; bh=d60x72jf42gLILD7IiqBL0OBb40=; b=iJd7eBugdIjzqW84UZ2xynlg1SojANJR6xfz0JDD44h78EpbqJiYVcMIfRG7mkrn741Bd5YaMR6p 9j41OA9A5am+8BE1Ng2kLRGwou5hRInn+xXBAQm2NUt5FkoXSpvm4gC4gQSOxPbQcuzlLha9JqxJ 8ZZ89\/20txUrRq9cYxk=", "v=1; a=rsa-sha256; c=relaxed\/relaxed; d=c.mandrillapp.com; i=@c.mandrillapp.com; q=dns\/txt; s=mandrill; t=1368214100; h=From : Sender : Subject : List-Unsubscribe : To : Message-Id : Date : MIME-Version : Content-Type : From : Subject : Date : X-Mandrill-User : List-Unsubscribe; bh=y5Vz+RDcKZmWzRc9s0xUJR6k4APvBNktBqy1EhSWM8o=; b=PLAUIuw8zk8kG5tPkmcnSanElxt6I5lp5t32nSvzVQE7R8u0AmIEjeIDZEt31+Q9PWt+nY sHHRsXUQ9SZpndT9Bk++\/SmyA2ntU\/2AKuqDpPkMZiTqxmGF80Wz4JJgx2aCEB1LeLVmFFwB 5Nr\/LBSlsBlRcjT9QiWw0\/iRvCn74="],
//             "Domainkey-Signature": "a=rsa-sha1; c=nofws; q=dns; s=mandrill; d=mail115.us4.mandrillapp.com; b=X6qudHd4oOJvVQZcoAEUCJgB875SwzEO5UKf6NvpfqyCVjdaO79WdDulLlfNVELeuoK2m6alt2yw 5Qhp4TW5NegyFf6Ogr\/Hy0Lt411r\/0lRf0nyaVkqMM\/9g13B6D9CS092v70wshX8+qdyxK8fADw8 kEelbCK2cEl0AGIeAeo=;",
//             "From": "<example.sender@mandrillapp.com>",
//             "List-Unsubscribe": "<mailto:unsubscribe-md_999.aaaaaaaa.v1-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@mailin1.us2.mcsv.net?subject=unsub>",
//             "Message-Id": "<999.20130510192820.aaaaaaaaaaaaaa.aaaaaaaa@mail115.us4.mandrillapp.com>",
//             "Mime-Version": "1.0",
//             "Received": ["from mail115.us4.mandrillapp.com (mail115.us4.mandrillapp.com [205.201.136.115]) by mail.example.com (Postfix) with ESMTP id AAAAAAAAAAA for <example@pinged.email>; Fri, 10 May 2013 19:28:21 +0000 (UTC)", "from localhost (127.0.0.1) by mail115.us4.mandrillapp.com id hhl55a14i282 for <example@pinged.email>; Fri, 10 May 2013 19:28:20 +0000 (envelope-from <bounce-md_999.aaaaaaaa.v1-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@mail115.us4.mandrillapp.com>)"],
//             "Sender": "<example.sender@mail115.us4.mandrillapp.com>",
//             "Subject": "This is an example webhook message",
//             "To": "<example@pinged.email>",
//             "X-Report-Abuse": "Please forward a copy of this message, including all headers, to abuse@mandrill.com"
//         },
//         "html": "<p>This is an example inbound message.<\/p><img src=\"http:\/\/mandrillapp.com\/track\/open.php?u=999&id=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa&tags=_all,_sendexample.sender@mandrillapp.com\" height=\"1\" width=\"1\">\n",
//         "raw_msg": "Received: from mail115.us4.mandrillapp.com (mail115.us4.mandrillapp.com [205.201.136.115])\n\tby mail.example.com (Postfix) with ESMTP id AAAAAAAAAAA\n\tfor <example@pinged.email>; Fri, 10 May 2013 19:28:20 +0000 (UTC)\nDKIM-Signature: v=1; a=rsa-sha1; c=relaxed\/relaxed; s=mandrill; d=mail115.us4.mandrillapp.com;\n h=From:Sender:Subject:List-Unsubscribe:To:Message-Id:Date:MIME-Version:Content-Type; i=example.sender@mail115.us4.mandrillapp.com;\n bh=d60x72jf42gLILD7IiqBL0OBb40=;\n b=iJd7eBugdIjzqW84UZ2xynlg1SojANJR6xfz0JDD44h78EpbqJiYVcMIfRG7mkrn741Bd5YaMR6p\n 9j41OA9A5am+8BE1Ng2kLRGwou5hRInn+xXBAQm2NUt5FkoXSpvm4gC4gQSOxPbQcuzlLha9JqxJ\n 8ZZ89\/20txUrRq9cYxk=\nDomainKey-Signature: a=rsa-sha1; c=nofws; q=dns; s=mandrill; d=mail115.us4.mandrillapp.com;\n b=X6qudHd4oOJvVQZcoAEUCJgB875SwzEO5UKf6NvpfqyCVjdaO79WdDulLlfNVELeuoK2m6alt2yw\n 5Qhp4TW5NegyFf6Ogr\/Hy0Lt411r\/0lRf0nyaVkqMM\/9g13B6D9CS092v70wshX8+qdyxK8fADw8\n kEelbCK2cEl0AGIeAeo=;\nReceived: from localhost (127.0.0.1) by mail115.us4.mandrillapp.com id hhl55a14i282 for <example@pinged.email>; Fri, 10 May 2013 19:28:20 +0000 (envelope-from <bounce-md_999.aaaaaaaa.v1-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@mail115.us4.mandrillapp.com>)\nDKIM-Signature: v=1; a=rsa-sha256; c=relaxed\/relaxed; d=c.mandrillapp.com; \n i=@c.mandrillapp.com; q=dns\/txt; s=mandrill; t=1368214100; h=From : \n Sender : Subject : List-Unsubscribe : To : Message-Id : Date : \n MIME-Version : Content-Type : From : Subject : Date : X-Mandrill-User : \n List-Unsubscribe; bh=y5Vz+RDcKZmWzRc9s0xUJR6k4APvBNktBqy1EhSWM8o=; \n b=PLAUIuw8zk8kG5tPkmcnSanElxt6I5lp5t32nSvzVQE7R8u0AmIEjeIDZEt31+Q9PWt+nY\n sHHRsXUQ9SZpndT9Bk++\/SmyA2ntU\/2AKuqDpPkMZiTqxmGF80Wz4JJgx2aCEB1LeLVmFFwB\n 5Nr\/LBSlsBlRcjT9QiWw0\/iRvCn74=\nFrom: <example.sender@mandrillapp.com>\nSender: <example.sender@mail115.us4.mandrillapp.com>\nSubject: This is an example webhook message\nList-Unsubscribe: <mailto:unsubscribe-md_999.aaaaaaaa.v1-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@mailin1.us2.mcsv.net?subject=unsub>\nTo: <example@pinged.email>\nX-Report-Abuse: Please forward a copy of this message, including all headers, to abuse@mandrill.com\nX-Mandrill-User: md_999\nMessage-Id: <999.20130510192820.aaaaaaaaaaaaaa.aaaaaaaa@mail115.us4.mandrillapp.com>\nDate: Fri, 10 May 2013 19:28:20 +0000\nMIME-Version: 1.0\nContent-Type: multipart\/alternative; boundary=\"_av-7r7zDhHxVEAo2yMWasfuFw\"\n\n--_av-7r7zDhHxVEAo2yMWasfuFw\nContent-Type: text\/plain; charset=utf-8\nContent-Transfer-Encoding: 7bit\n\nThis is an example inbound message.\n--_av-7r7zDhHxVEAo2yMWasfuFw\nContent-Type: text\/html; charset=utf-8\nContent-Transfer-Encoding: 7bit\n\n<p>This is an example inbound message.<\/p><img src=\"http:\/\/mandrillapp.com\/track\/open.php?u=999&id=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa&tags=_all,_sendexample.sender@mandrillapp.com\" height=\"1\" width=\"1\">\n--_av-7r7zDhHxVEAo2yMWasfuFw--",
//         "sender": null,
//         "spam_report": {
//             "matched_rules": [{
//                 "description": "RBL: Sender listed at http:\/\/www.dnswl.org\/, no",
//                 "name": "RCVD_IN_DNSWL_NONE",
//                 "score": 0
//             }, {
//                 "description": null,
//                 "name": null,
//                 "score": 0
//             }, {
//                 "description": "in iadb.isipp.com]",
//                 "name": "listed",
//                 "score": 0
//             }, {
//                 "description": "RBL: Participates in the IADB system",
//                 "name": "RCVD_IN_IADB_LISTED",
//                 "score": -0.4
//             }, {
//                 "description": "RBL: ISIPP IADB lists as vouched-for sender",
//                 "name": "RCVD_IN_IADB_VOUCHED",
//                 "score": -2.2
//             }, {
//                 "description": "RBL: IADB: Sender publishes SPF record",
//                 "name": "RCVD_IN_IADB_SPF",
//                 "score": 0
//             }, {
//                 "description": "RBL: IADB: Sender publishes Sender ID record",
//                 "name": "RCVD_IN_IADB_SENDERID",
//                 "score": 0
//             }, {
//                 "description": "RBL: IADB: Sender publishes Domain Keys record",
//                 "name": "RCVD_IN_IADB_DK",
//                 "score": -0.2
//             }, {
//                 "description": "RBL: IADB: Sender has reverse DNS record",
//                 "name": "RCVD_IN_IADB_RDNS",
//                 "score": -0.2
//             }, {
//                 "description": "SPF: HELO matches SPF record",
//                 "name": "SPF_HELO_PASS",
//                 "score": 0
//             }, {
//                 "description": "BODY: HTML included in message",
//                 "name": "HTML_MESSAGE",
//                 "score": 0
//             }, {
//                 "description": "BODY: HTML: images with 0-400 bytes of words",
//                 "name": "HTML_IMAGE_ONLY_04",
//                 "score": 0.3
//             }, {
//                 "description": "Message has a DKIM or DK signature, not necessarily valid",
//                 "name": "DKIM_SIGNED",
//                 "score": 0.1
//             }, {
//                 "description": "Message has at least one valid DKIM or DK signature",
//                 "name": "DKIM_VALID",
//                 "score": -0.1
//             }],
//             "score": -2.6
//         },
//         "spf": {
//             "detail": "sender SPF authorized",
//             "result": "pass"
//         },
//         "subject": "This is an example webhook message",
//         "tags": [],
//         "template": null,
//         "text": "This is an example inbound message.\n",
//         "text_flowed": false,
//         "to": [
//             ["example@pinged.email", null]
//         ]
//     },
//     "ts": 1368214102
// }, {
//     "event": "inbound",
//     "msg": {
//         "dkim": {
//             "signed": true,
//             "valid": true
//         },
//         "email": "example@pinged.email",
//         "from_email": "example.sender@mandrillapp.com",
//         "headers": {
//             "Content-Type": "multipart\/alternative; boundary=\"_av-7r7zDhHxVEAo2yMWasfuFw\"",
//             "Date": "Fri, 10 May 2013 19:28:20 +0000",
//             "Dkim-Signature": ["v=1; a=rsa-sha1; c=relaxed\/relaxed; s=mandrill; d=mail115.us4.mandrillapp.com; h=From:Sender:Subject:List-Unsubscribe:To:Message-Id:Date:MIME-Version:Content-Type; i=example.sender@mail115.us4.mandrillapp.com; bh=d60x72jf42gLILD7IiqBL0OBb40=; b=iJd7eBugdIjzqW84UZ2xynlg1SojANJR6xfz0JDD44h78EpbqJiYVcMIfRG7mkrn741Bd5YaMR6p 9j41OA9A5am+8BE1Ng2kLRGwou5hRInn+xXBAQm2NUt5FkoXSpvm4gC4gQSOxPbQcuzlLha9JqxJ 8ZZ89\/20txUrRq9cYxk=", "v=1; a=rsa-sha256; c=relaxed\/relaxed; d=c.mandrillapp.com; i=@c.mandrillapp.com; q=dns\/txt; s=mandrill; t=1368214100; h=From : Sender : Subject : List-Unsubscribe : To : Message-Id : Date : MIME-Version : Content-Type : From : Subject : Date : X-Mandrill-User : List-Unsubscribe; bh=y5Vz+RDcKZmWzRc9s0xUJR6k4APvBNktBqy1EhSWM8o=; b=PLAUIuw8zk8kG5tPkmcnSanElxt6I5lp5t32nSvzVQE7R8u0AmIEjeIDZEt31+Q9PWt+nY sHHRsXUQ9SZpndT9Bk++\/SmyA2ntU\/2AKuqDpPkMZiTqxmGF80Wz4JJgx2aCEB1LeLVmFFwB 5Nr\/LBSlsBlRcjT9QiWw0\/iRvCn74="],
//             "Domainkey-Signature": "a=rsa-sha1; c=nofws; q=dns; s=mandrill; d=mail115.us4.mandrillapp.com; b=X6qudHd4oOJvVQZcoAEUCJgB875SwzEO5UKf6NvpfqyCVjdaO79WdDulLlfNVELeuoK2m6alt2yw 5Qhp4TW5NegyFf6Ogr\/Hy0Lt411r\/0lRf0nyaVkqMM\/9g13B6D9CS092v70wshX8+qdyxK8fADw8 kEelbCK2cEl0AGIeAeo=;",
//             "From": "<example.sender@mandrillapp.com>",
//             "List-Unsubscribe": "<mailto:unsubscribe-md_999.aaaaaaaa.v1-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@mailin1.us2.mcsv.net?subject=unsub>",
//             "Message-Id": "<999.20130510192820.aaaaaaaaaaaaaa.aaaaaaaa@mail115.us4.mandrillapp.com>",
//             "Mime-Version": "1.0",
//             "Received": ["from mail115.us4.mandrillapp.com (mail115.us4.mandrillapp.com [205.201.136.115]) by mail.example.com (Postfix) with ESMTP id AAAAAAAAAAA for <example@pinged.email>; Fri, 10 May 2013 19:28:21 +0000 (UTC)", "from localhost (127.0.0.1) by mail115.us4.mandrillapp.com id hhl55a14i282 for <example@pinged.email>; Fri, 10 May 2013 19:28:20 +0000 (envelope-from <bounce-md_999.aaaaaaaa.v1-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@mail115.us4.mandrillapp.com>)"],
//             "Sender": "<example.sender@mail115.us4.mandrillapp.com>",
//             "Subject": "This is an example webhook message",
//             "To": "<example@pinged.email>",
//             "X-Report-Abuse": "Please forward a copy of this message, including all headers, to abuse@mandrill.com"
//         },
//         "html": "<p>This is an example inbound message.<\/p><img src=\"http:\/\/mandrillapp.com\/track\/open.php?u=999&id=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa&tags=_all,_sendexample.sender@mandrillapp.com\" height=\"1\" width=\"1\">\n",
//         "raw_msg": "Received: from mail115.us4.mandrillapp.com (mail115.us4.mandrillapp.com [205.201.136.115])\n\tby mail.example.com (Postfix) with ESMTP id AAAAAAAAAAA\n\tfor <example@pinged.email>; Fri, 10 May 2013 19:28:20 +0000 (UTC)\nDKIM-Signature: v=1; a=rsa-sha1; c=relaxed\/relaxed; s=mandrill; d=mail115.us4.mandrillapp.com;\n h=From:Sender:Subject:List-Unsubscribe:To:Message-Id:Date:MIME-Version:Content-Type; i=example.sender@mail115.us4.mandrillapp.com;\n bh=d60x72jf42gLILD7IiqBL0OBb40=;\n b=iJd7eBugdIjzqW84UZ2xynlg1SojANJR6xfz0JDD44h78EpbqJiYVcMIfRG7mkrn741Bd5YaMR6p\n 9j41OA9A5am+8BE1Ng2kLRGwou5hRInn+xXBAQm2NUt5FkoXSpvm4gC4gQSOxPbQcuzlLha9JqxJ\n 8ZZ89\/20txUrRq9cYxk=\nDomainKey-Signature: a=rsa-sha1; c=nofws; q=dns; s=mandrill; d=mail115.us4.mandrillapp.com;\n b=X6qudHd4oOJvVQZcoAEUCJgB875SwzEO5UKf6NvpfqyCVjdaO79WdDulLlfNVELeuoK2m6alt2yw\n 5Qhp4TW5NegyFf6Ogr\/Hy0Lt411r\/0lRf0nyaVkqMM\/9g13B6D9CS092v70wshX8+qdyxK8fADw8\n kEelbCK2cEl0AGIeAeo=;\nReceived: from localhost (127.0.0.1) by mail115.us4.mandrillapp.com id hhl55a14i282 for <example@pinged.email>; Fri, 10 May 2013 19:28:20 +0000 (envelope-from <bounce-md_999.aaaaaaaa.v1-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@mail115.us4.mandrillapp.com>)\nDKIM-Signature: v=1; a=rsa-sha256; c=relaxed\/relaxed; d=c.mandrillapp.com; \n i=@c.mandrillapp.com; q=dns\/txt; s=mandrill; t=1368214100; h=From : \n Sender : Subject : List-Unsubscribe : To : Message-Id : Date : \n MIME-Version : Content-Type : From : Subject : Date : X-Mandrill-User : \n List-Unsubscribe; bh=y5Vz+RDcKZmWzRc9s0xUJR6k4APvBNktBqy1EhSWM8o=; \n b=PLAUIuw8zk8kG5tPkmcnSanElxt6I5lp5t32nSvzVQE7R8u0AmIEjeIDZEt31+Q9PWt+nY\n sHHRsXUQ9SZpndT9Bk++\/SmyA2ntU\/2AKuqDpPkMZiTqxmGF80Wz4JJgx2aCEB1LeLVmFFwB\n 5Nr\/LBSlsBlRcjT9QiWw0\/iRvCn74=\nFrom: <example.sender@mandrillapp.com>\nSender: <example.sender@mail115.us4.mandrillapp.com>\nSubject: This is an example webhook message\nList-Unsubscribe: <mailto:unsubscribe-md_999.aaaaaaaa.v1-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@mailin1.us2.mcsv.net?subject=unsub>\nTo: <example@pinged.email>\nX-Report-Abuse: Please forward a copy of this message, including all headers, to abuse@mandrill.com\nX-Mandrill-User: md_999\nMessage-Id: <999.20130510192820.aaaaaaaaaaaaaa.aaaaaaaa@mail115.us4.mandrillapp.com>\nDate: Fri, 10 May 2013 19:28:20 +0000\nMIME-Version: 1.0\nContent-Type: multipart\/alternative; boundary=\"_av-7r7zDhHxVEAo2yMWasfuFw\"\n\n--_av-7r7zDhHxVEAo2yMWasfuFw\nContent-Type: text\/plain; charset=utf-8\nContent-Transfer-Encoding: 7bit\n\nThis is an example inbound message.\n--_av-7r7zDhHxVEAo2yMWasfuFw\nContent-Type: text\/html; charset=utf-8\nContent-Transfer-Encoding: 7bit\n\n<p>This is an example inbound message.<\/p><img src=\"http:\/\/mandrillapp.com\/track\/open.php?u=999&id=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa&tags=_all,_sendexample.sender@mandrillapp.com\" height=\"1\" width=\"1\">\n--_av-7r7zDhHxVEAo2yMWasfuFw--",
//         "sender": null,
//         "spam_report": {
//             "matched_rules": [{
//                 "description": "RBL: Sender listed at http:\/\/www.dnswl.org\/, no",
//                 "name": "RCVD_IN_DNSWL_NONE",
//                 "score": 0
//             }, {
//                 "description": null,
//                 "name": null,
//                 "score": 0
//             }, {
//                 "description": "in iadb.isipp.com]",
//                 "name": "listed",
//                 "score": 0
//             }, {
//                 "description": "RBL: Participates in the IADB system",
//                 "name": "RCVD_IN_IADB_LISTED",
//                 "score": -0.4
//             }, {
//                 "description": "RBL: ISIPP IADB lists as vouched-for sender",
//                 "name": "RCVD_IN_IADB_VOUCHED",
//                 "score": -2.2
//             }, {
//                 "description": "RBL: IADB: Sender publishes SPF record",
//                 "name": "RCVD_IN_IADB_SPF",
//                 "score": 0
//             }, {
//                 "description": "RBL: IADB: Sender publishes Sender ID record",
//                 "name": "RCVD_IN_IADB_SENDERID",
//                 "score": 0
//             }, {
//                 "description": "RBL: IADB: Sender publishes Domain Keys record",
//                 "name": "RCVD_IN_IADB_DK",
//                 "score": -0.2
//             }, {
//                 "description": "RBL: IADB: Sender has reverse DNS record",
//                 "name": "RCVD_IN_IADB_RDNS",
//                 "score": -0.2
//             }, {
//                 "description": "SPF: HELO matches SPF record",
//                 "name": "SPF_HELO_PASS",
//                 "score": 0
//             }, {
//                 "description": "BODY: HTML included in message",
//                 "name": "HTML_MESSAGE",
//                 "score": 0
//             }, {
//                 "description": "BODY: HTML: images with 0-400 bytes of words",
//                 "name": "HTML_IMAGE_ONLY_04",
//                 "score": 0.3
//             }, {
//                 "description": "Message has a DKIM or DK signature, not necessarily valid",
//                 "name": "DKIM_SIGNED",
//                 "score": 0.1
//             }, {
//                 "description": "Message has at least one valid DKIM or DK signature",
//                 "name": "DKIM_VALID",
//                 "score": -0.1
//             }],
//             "score": -2.6
//         },
//         "spf": {
//             "detail": "sender SPF authorized",
//             "result": "pass"
//         },
//         "subject": "This is an example webhook message",
//         "tags": [],
//         "template": null,
//         "text": "This is an example inbound message.\n",
//         "text_flowed": false,
//         "to": [
//             ["example@pinged.email", null]
//         ]
//     },
//     "ts": 1368214102
// }]
