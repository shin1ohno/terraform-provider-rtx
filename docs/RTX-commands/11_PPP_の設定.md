# 第11章: PPP の設定

> 元PDFページ: 218-240

---

第 11 章
PPP の設定
11.1 相手の名前とパスワードの設定
[書式 ]
pp auth  username  username  password  [myname myname  mypass ] [isdn1 ] [clid [ isdn2 ...]] [mscbcp] [ ip_address ]
[ip6_prefix ]
pp auth  username  username  password  [myname myname  mypass ] [ip_address ] [ip6_prefix ]
no pp auth  username  username  [password ...]
[設定値及び初期値 ]
•username
• [設定値 ] : 名前  (64 文字以内  )
• [初期値 ] : -
•password
• [設定値 ] : パスワード  (64 文字以内  )
• [初期値 ] : -
• myname : 自分側の設定を入力するためのキーワード
• [初期値 ] : -
•myname
• [設定値 ] : 自分側のユーザ名
• [初期値 ] : -
•mypass
• [設定値 ] : 自分側のパスワード
• [初期値 ] : -
•isdn1
• [設定値 ] : 相手の  ISDN アドレス
• [初期値 ] : -
• clid : 発番号認証を利用することを示すキーワード
• [初期値 ] : -
•isdn2
• [設定値 ] : 発番号認証に用いられる  ISDN アドレス
• [初期値 ] : -
• mscbcp : MS コールバックを許可することを示すキーワード
• [初期値 ] : -
•ip_address
• [設定値 ] : 相手に割り当てる  IP アドレス
• [初期値 ] : -
•ip6_prefix
• [設定値 ] : ユーザに割り当てるプレフィックス
• [初期値 ] : -
[説明 ]
相手の名前とパスワードを設定する。複数の設定が可能。
オプションで自分側の設定も入力ができる。
BRI インタフェースを持たないモデルでは第  2 書式を用いる。
双方向で認証を行う場合には、相手のユーザ名が確定してから自分を相手に認証させるプロセスが動き始める。
これらのパラメータが設定されていない場合には、  pp auth myname  コマンドの設定が参照される。218 | コマンドリファレンス  | PPP の設定

オプションで  ISDN 番号が設定でき、名前と結びついたルーティングやリモート  IP アドレスに対しての発信を可能
にする。  isdn1  は発信用の  ISDN アドレスである。 isdn1  を省略すると、この相手には発信しなくなる。
名前に  '*' を与えた場合にはワイルドカードとして扱い、他の名前とマッチしなかった相手に対してその設定を使用
する。
clid キーワードは発番号認証を利用することを指示する。このキーワードがない場合は発番号認証は行われない。
発番号認証は isdn2  があれば isdn2  を用い、または isdn2  がなければ isdn1  を用い、一致したら認証は成功したとみな
す。
isdn2  は複数設定することができる。複数設定する場合は、まず先頭の  ISDN アドレスで認証が行われ、認証に失敗
すると次の  ISDN アドレスが使われる。
mscbcp キーワードは  MS コールバックを許可することを指示する。このユーザからの着信に対しては、同時に isdn
callback permit  on としてあれば  MS コールバックの動作を行う。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.2 受け入れる認証タイプの設定
[書式 ]
pp auth  accept  accept  [accept ]
no pp auth  accept  [accept ]
[設定値及び初期値 ]
•accept
• [設定値 ] :
設定値 説明
pap PAP による認証を受け入れる
chap CHAP による認証を受け入れる
mschap MSCHAP による認証を受け入れる
mschap-v2 MSCHAP Version2 による認証を受け入れる
• [初期値 ] : 認証を受け入れない
[説明 ]
相手からの  PPP 認証要求を受け入れるかどうか設定する。発信時には常に適用される。 anonymous でない着信の場
合には発番号により  PP が選択されてから適用される。 anonymous での着信時には、発番号による  PP の選択が失敗
した場合に適用される。
このコマンドで認証を受け入れる設定になっていても、 pp auth myname  コマンドで自分の名前とパスワードが設定
されていなければ、認証を拒否する。
PP 毎のコマンドである。
[ノート ]
auth パラメーターへの  mschap、mschap-v2 の指定は、 RTX5000 / RTX3500 Rev.14.00.11 以前では不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.3 要求する認証タイプの設定
[書式 ]
pp auth  request  auth [arrive-only]
no pp auth  request  [auth[arrive-only]]
[設定値及び初期値 ]
•authコマンドリファレンス  | PPP の設定  | 219

• [設定値 ] :
設定値 説明
pap PAP による認証を要求する
chap CHAP による認証を要求する
mschap MSCHAP による認証を要求する
mschap-v2 MSCHAP Version2 による認証を要求する
chap-pap CHAP もしくは  PAP による認証を要求する
• [初期値 ] : -
[説明 ]
選択された相手について  PAP と CHAP による認証を要求するかどうかを設定する。発信時には常に適用される。
anonymous でない着信の場合には発番号により  PP が選択されてから適用される。 anonymous での着信時には、発番
号による  PP の選択が失敗した場合に適用される。
chap-pap キーワードの場合には、最初  CHAP を要求し、それが相手から拒否された場合には改めて  PAP を要求する
よう動作する。これにより、相手が  PAP または  CHAP の片方しかサポートしていない場合でも容易に接続できるよ
うになる。
arrive-only キーワードが指定された場合には、着信時にのみ  PPP による認証を要求するようになり、発信時には要
求しない。
[ノート ]
auth パラメーターへの  mschap、mschap-v2 の指定は、 RTX5000 / RTX3500 Rev.14.00.11 以前では不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.4 自分の名前とパスワードの設定
[書式 ]
pp auth  myname  myname  password
no pp auth  myname  [myname  password ]
[設定値及び初期値 ]
•myname
• [設定値 ] : 名前  (64 文字以内  )
• [初期値 ] : -
•password
• [設定値 ] : パスワード  (64 文字以内  )
• [初期値 ] : -
[説明 ]
PAP または  CHAP で相手に送信する自分の名前とパスワードを設定する。
PP 毎のコマンドである。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.5 同一  username を持つ相手からの二重接続を禁止するか否かの設定
[書式 ]
pp auth  multi  connect  prohibit  prohibit
no pp auth  multi  connect  prohibit  [prohibit ]
[設定値及び初期値 ]
•prohibit220 | コマンドリファレンス  | PPP の設定

• [設定値 ] :
設定値 説明
on 禁止する
off 禁止しない
• [初期値 ] : off
[説明 ]
pp auth username  コマンドで登録した同一 username  を持つ相手からの二重接続を禁止するか否かを設定する。
[ノート ]
定額制プロバイダを営む場合に便利である。ユーザ管理を  RADIUS で行う場合には、二重接続の禁止は  RADIUS サ
ーバーの方で対処する必要がある。
anonymous が選択された場合のみ有効である。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.6 LCP 関連の設定
11.6.1 Address and Control Field Compression オプション使用の設定
[書式 ]
ppp lcp acfc acfc
no ppp lcp acfc [acfc]
[設定値及び初期値 ]
•acfc
• [設定値 ] :
設定値 説明
on 用いる
off 用いない
• [初期値 ] : off
[説明 ]
選択されている相手について [PPP,LCP] の Address and Control Field Compression オプションを用いるか否かを設定す
る。
[ノート ]
on を設定していても相手に拒否された場合は用いない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.6.2 Magic Number オプション使用の設定
[書式 ]
ppp lcp magicnumber  magicnumber
no ppp lcp magicnumber  [magicnumber ]
[設定値及び初期値 ]
•magicnumber
• [設定値 ] :
設定値 説明
on 用いる
off 用いないコマンドリファレンス  | PPP の設定  | 221

• [初期値 ] : on
[説明 ]
選択されている相手について [PPP,LCP] の Magic Number オプションを用いるか否かを設定する。
[ノート ]
on を設定していても相手に拒否された場合は用いない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.6.3 Maximum Receive Unit オプション使用の設定
[書式 ]
ppp lcp mru mru [length ]
no ppp lcp mru [mru[length ]]
[設定値及び初期値 ]
•mru
• [設定値 ] :
設定値 説明
on 用いる
off 用いない
• [初期値 ] : on
•length  : MRU の値
• [設定値 ] :
• 1280..1792
• [初期値 ] : 1792
[説明 ]
選択されている相手について [PPP,LCP] の Maximum Receive Unit オプションを用いるか否かと、 MRU の値を設定す
る。
[ノート ]
on を設定していても相手に拒否された場合は用いない。一般には  on でよいが、 このオプションをつけると接続でき
ないルーターに接続する場合には  off にする。
データ圧縮を利用する設定の場合には、 length  パラメータの設定は常に  1792 として動作する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.6.4 Protocol Field Compression オプション使用の設定
[書式 ]
ppp lcp pfc pfc
no ppp lcp pfc [pfc]
[設定値及び初期値 ]
•pfc
• [設定値 ] :
設定値 説明
on 用いる
off 用いない
• [初期値 ] : off
[説明 ]
選択されている相手について [PPP,LCP] の Protocol Field Compression オプションを用いるか否かを設定する。222 | コマンドリファレンス  | PPP の設定

[ノート ]
on を設定していても相手に拒否された場合は用いない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.6.5 lcp-restart パラメータの設定
[書式 ]
ppp lcp restart  time
no ppp lcp restart  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : ミリ秒  (20..10000)
• [初期値 ] : 3000
[説明 ]
選択されている相手について [PPP,LCP] の configure-request 、terminate-request の再送時間を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.6.6 lcp-max-terminate パラメータの設定
[書式 ]
ppp lcp maxterminate  count
no ppp lcp maxterminate  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 2
[説明 ]
選択されている相手について [PPP,LCP] の terminate-request の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.6.7 lcp-max-configure パラメータの設定
[書式 ]
ppp lcp maxconfigure  count
no ppp lcp maxconfigure  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 10
[説明 ]
選択されている相手について [PPP,LCP] の configure-request の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.6.8 lcp-max-failure パラメータの設定
[書式 ]
ppp lcp maxfailure  count
no ppp lcp maxfailure  [count ]コマンドリファレンス  | PPP の設定  | 223

[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 10
[説明 ]
選択されている相手について [PPP,LCP] の configure-nak の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.6.9 Configure-Request をすぐに送信するか否かの設定
[書式 ]
ppp lcp silent  switch
no ppp lcp silent  [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on PPP/LCP で、回線接続直後の  Configure-Request の送
信を、相手から  Configure-Request を受信するまで遅
らせる
off PPP/LCP で、回線接続直後に  Configure-Request を送
信する
• [初期値 ] : off
[説明 ]
PPP/LCP で、回線接続後  Configure-Request をすぐに送信するか、あるいは相手から  Configure-Request を受信するま
で遅らせるかを設定する。通常は回線接続直後に  Configure-Request を送信して構わないが、接続相手によってはこ
れを遅らせた方がよいものがある。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.7 PAP 関連の設定
11.7.1 pap-restart パラメータの設定
[書式 ]
ppp pap restart  time
no ppp pap restart  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : ミリ秒  (20..10000)
• [初期値 ] : 3000
[説明 ]
選択されている相手について [PPP,PAP]authenticate-request の再送時間を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.7.2 pap-max-authreq パラメータの設定
[書式 ]
ppp pap maxauthreq  count
no ppp pap maxauthreq  [count ]224 | コマンドリファレンス  | PPP の設定

[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 10
[説明 ]
選択されている相手について [PPP,PAP]authenticate-request の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.8 CHAP 関連の設定
11.8.1 chap-restart パラメータの設定
[書式 ]
ppp chap  restart  time
no ppp chap  restart  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : ミリ秒  (20..10000)
• [初期値 ] : 3000
[説明 ]
選択されている相手について [PPP,CHAP]challenge の再送時間を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.8.2 chap-max-challenge パラメータの設定
[書式 ]
ppp chap  maxchallenge  count
no ppp chap  maxchallenge  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 10
[説明 ]
選択されている相手について [PPP,CHAP]challenge の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.9 IPCP 関連の設定
11.9.1 Van Jacobson Compressed TCP/IP 使用の設定
[書式 ]
ppp ipcp vjc compression
no ppp ipcp vjc [compression ]
[設定値及び初期値 ]
•compression
• [設定値 ] :
設定値 説明
on 使用するコマンドリファレンス  | PPP の設定  | 225

設定値 説明
off 使用しない
• [初期値 ] : off
[説明 ]
選択されている相手について [PPP,IPCP]Van Jacobson Compressed TCP/IP を使用するか否かを設定する。
[ノート ]
on を設定していても相手に拒否された場合は用いない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.9.2 PP 側 IP アドレスのネゴシエーションの設定
[書式 ]
ppp ipcp ipaddress  negotiation
no ppp ipcp ipaddress  [negotiation ]
[設定値及び初期値 ]
•negotiation
• [設定値 ] :
設定値 説明
on ネゴシエーションする
off ネゴシエーションしない
• [初期値 ] : off
[説明 ]
選択されている相手について  PP 側 IP アドレスのネゴシエーションをするか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.9.3 ipcp-restart パラメータの設定
[書式 ]
ppp ipcp restart  time
no ppp ipcp restart  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : ミリ秒  (20..10000)
• [初期値 ] : 3000
[説明 ]
選択されている相手について [PPP,IPCP] の configure-request 、terminate-request の再送時間を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.9.4 ipcp-max-terminate パラメータの設定
[書式 ]
ppp ipcp maxterminate  count
no ppp ipcp maxterminate  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 2226 | コマンドリファレンス  | PPP の設定

[説明 ]
選択されている相手について [PPP,IPCP] の terminate-request の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.9.5 ipcp-max-configure パラメータの設定
[書式 ]
ppp ipcp maxconfigure  count
no ppp ipcp maxconfigure  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 10
[説明 ]
選択されている相手について [PPP,IPCP] の configure-request の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.9.6 ipcp-max-failure パラメータの設定
[書式 ]
ppp ipcp maxfailure  count
no ppp ipcp maxfailure  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 10
[説明 ]
選択されている相手について [PPP,IPCP] の configure-nak の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.9.7 WINS サーバーの  IP アドレスの設定
[書式 ]
wins  server  server1  [server2 ]
no wins  server  [server1  [server2 ]]
[設定値及び初期値 ]
•server1、server2
• [設定値 ] : IP アドレス  (xxx.xxx.xxx.xxx(xxx は十進数  ))
• [初期値 ] : -
[説明 ]
WINS(Windows Internet Name Service) サーバーの  IP アドレスを設定する。
[ノート ]
IPCP の MS 拡張オプションおよび  DHCP でクライアントに渡すための  WINS サーバーの  IP アドレスを設定する。
ルーターはこのサーバーに対し  WINS クライアントとしての動作は一切行わない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | PPP の設定  | 227

11.9.8 IPCP の MS 拡張オプションを使うか否かの設定
[書式 ]
ppp ipcp msext  msext
no ppp ipcp msext  [msext ]
[設定値及び初期値 ]
•msext
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : off
[説明 ]
選択されている相手について、 [PPP,IPCP] の MS 拡張オプションを使うか否かを設定する。
IPCP の Microsoft 拡張オプションを使うように設定すると、 DNS サーバーの  IP アドレスと  WINS(Windows Internet
Name Service) サーバーの  IP アドレスを、接続した相手である  Windows マシンに渡すことができる。渡すための
DNS サーバーや  WINS サーバーの  IP アドレスはそれぞれ、 dns server  コマンドおよび wins server  コマンドで設定す
る。
off の場合は、 DNS サーバーや  WINS サーバーのアドレスを渡されても受け取らない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.9.9 ホスト経路が存在する相手側  IP アドレスを受け入れるか否かの設定
[書式 ]
ppp ipcp remote  address  check  sw
no ppp ipcp remote  address  check  [sw]
[設定値及び初期値 ]
•sw
• [設定値 ] :
設定値 説明
on 通知された相手の  PP 側 IP アドレスを拒否する
off 通知された相手の  PP 側 IP アドレスを受け入れる
• [初期値 ] : on
[説明 ]
他の  PP 経由のホスト経路が既に存在している  IP アドレスを  PP 接続時に相手側  IP アドレスとして通知されたとき
に、その  IP アドレスを受け入れるか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.10 MSCBCP 関連の設定
11.10.1 mscbcp-restart パラメータの設定
[書式 ]
ppp mscbcp  restart  time
no ppp mscbcp  restart  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : ミリ秒  (20..10000)228 | コマンドリファレンス  | PPP の設定

• [初期値 ] : 1000
[説明 ]
選択されている相手について [PPP,MSCBCP] の request/Response の再送時間を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.10.2 mscbcp-maxretry パラメータの設定
[書式 ]
ppp mscbcp  maxretry  count
no ppp mscbcp  maxretry  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..30)
• [初期値 ] : 30
[説明 ]
選択されている相手について [PPP,MSCBCP] の request/Response の再送回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.11 CCP 関連の設定
11.11.1 全パケットの圧縮タイプの設定
[書式 ]
ppp ccp type type
no ppp ccp type [type]
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
stac0 Stac LZS で圧縮する
stac Stac LZS で圧縮する
cstac Stac LZS で圧縮する  ( 接続相手が  Cisco ルーターの
場合  )
mppe-40 40bit MPPE で暗号化する
mppe-128 128bit MPPE で暗号化する
mppe-any 40bit,128bit MPPE いずれかの暗号化を行う
none 圧縮しない
• [初期値 ] : stac
[説明 ]
選択されている相手について [PPP,CCP] 圧縮方式を選択する。
[ノート ]
Van Jacobson Compressed TCP/IP との併用も可能である。
type に stac を指定した時、回線状態が悪い場合や、高負荷で、パケットロスが頻繁に起きると、通信が正常に行え
なくなることがある。このような場合、自動的に「圧縮なし」になる。その後、リスタートまで「圧縮なし」のま
まである。このような状況が改善できない時は、 stac0 を指定すればよい。ただしその時は接続先も  stac0 に対応し
ていなければならない。 stac0 は stac よりも圧縮効率は落ちる。コマンドリファレンス  | PPP の設定  | 229

接続相手が  Cisco ルーターの場合に  stac を適用すると通信できないことがある。そのような場合には、設定を  cstac
に変更すると通信が可能になることがある。
mppe-40,mppe-128,mppe-any の場合には  1 パケット毎に鍵交換される。 MPPE は Microsoft Point-To-Point
Encryption(Protocol) の略で  CCP を拡張したものであり、暗号アルゴリズムとして  RC4 を採用し、鍵長  40bit または
128bit を使う。暗号鍵生成のために認証プロトコルの  MS-CHAP または  MS-CHAPv2 と合わせて設定する。
vRX シリーズ、 RTX5000 、RTX3510 、RTX3500 では  stac0,stac,cstac,none の指定が可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.11.2 ccp-restart パラメータの設定
[書式 ]
ppp ccp restart  time
no ppp ccp restart  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : ミリ秒  (20..10000)
• [初期値 ] : 3000
[説明 ]
選択されている相手について [PPP,CCP] の configure-request 、terminate-request の再送時間を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.11.3 ccp-max-terminate パラメータの設定
[書式 ]
ppp ccp maxterminate  count
no ppp ccp maxterminate  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 2
[説明 ]
選択されている相手について [PPP,CCP] の terminate-request の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.11.4 ccp-max-configure パラメータの設定
[書式 ]
ppp ccp maxconfigure  count
no ppp ccp maxconfigure  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 10
[説明 ]
選択されている相手について [PPP,CCP] の configure-request の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830230 | コマンドリファレンス  | PPP の設定

11.11.5 ccp-max-failure パラメータの設定
[書式 ]
ppp ccp maxfailure  count
no ppp ccp maxfailure  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 10
[説明 ]
選択されている相手について [PPP,CCP] の configure-nak の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.12 IPV6CP 関連の設定
11.12.1 IPV6CP を使用するか否かの設定
[書式 ]
ppp ipv6cp  use use
no ppp ipv6cp  use [use]
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : on
[説明 ]
選択されている相手について  IPV6CP を使用するか否かを選択する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.13 MP 関連の設定
11.13.1 MP を使用するか否かの設定
[書式 ]
ppp mp use use
no ppp mp use [use]
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : offコマンドリファレンス  | PPP の設定  | 231

[説明 ]
選択されている相手について  MP を使用するか否かを選択する。
on に設定していても、 LCP の段階で相手とのネゴシエーションが成立しなければ  MP を使わずに通信する。
[適用モデル ]
RTX5000, RTX3500, RTX1210
11.13.2 MP の制御方法の設定
[書式 ]
ppp mp control  type
no ppp mp control  [type]
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
arrive 自分が  1B 目の着信側の場合に  MP を制御する
both 自分が  1B 目の発信着信いずれの場合でも  MP を制
御する
call 自分が  1B 目の発信側の場合に  MP を制御する
• [初期値 ] : call
[説明 ]
選択されている相手について  MP を制御して  2B 目の発信 /切断を行う場合を設定する。通常は初期値のように自分
が 1B 目の発信側の場合だけ制御するようにしておく。
[適用モデル ]
RTX5000, RTX3500, RTX1210
11.13.3 MP のための負荷閾値の設定
[書式 ]
ppp mp load threshold  call_load  call_count  disc_load  disc_count
no ppp mp load threshold  [call_load  call_count  disc_load  disc_count ]
[設定値及び初期値 ]
•call_load
• [設定値 ] : 発信負荷閾値  %(1..100)
• [初期値 ] : 70
•call_count
• [設定値 ] : 回数  (1..100)
• [初期値 ] : 1
•disc_load
• [設定値 ] : 切断負荷閾値  %(0..50)
• [初期値 ] : 30
•disc_count
• [設定値 ] : 回数  (1..100)
• [初期値 ] : 2
[説明 ]
選択されている相手について [PPP,MP] の 2B 目を発信したり切断したりする場合のデータ転送負荷の閾値を設定す
る。
負荷は回線速度に対する  % で評価し、送受信で大きい方の値を採用する。 call_load  を超える負荷が call_count  回繰
り返されたら  2B 目の発信を行う。逆に disc_load  を下回る負荷が disc_count  回繰り返されたら  2B 目を切断する。232 | コマンドリファレンス  | PPP の設定

[適用モデル ]
RTX5000, RTX3500, RTX1210
11.13.4 MP の最大リンク数の設定
[書式 ]
ppp mp maxlink  number
no ppp mp maxlink  [number ]
[設定値及び初期値 ]
•number
• [設定値 ] : リンク数
• [初期値 ] : 2
[説明 ]
選択されている相手について [PPP,MP] の最大リンク数を設定する。リンク数の最大値は、使用モデルで使用できる
ISDN Bch の数までとなる。
[適用モデル ]
RTX5000, RTX3500, RTX1210
11.13.5 MP の最小リンク数の設定
[書式 ]
ppp mp minlink  number
no ppp mp minlink  [number ]
[設定値及び初期値 ]
•number
• [設定値 ] : リンク数
• [初期値 ] : 1
[説明 ]
選択されている相手について [PPP,MP] の最小リンク数を設定する。
[適用モデル ]
RTX5000, RTX3500, RTX1210
11.13.6 MP のための負荷計測間隔の設定
[書式 ]
ppp mp timer  time
no ppp mp timer  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : 秒数  (1..21474836)
• [初期値 ] : 10
[説明 ]
選択されている相手について [PPP,MP] のための負荷計測間隔を設定する。
単位は秒。負荷計測だけでなく、すべての  MP の動作はこのコマンドで設定した間隔で行われる。
[適用モデル ]
RTX5000, RTX3500, RTX1210
11.13.7 MP のパケットを分割するか否かの設定
[書式 ]
ppp mp divide  divide
no ppp mp divide  [divide ]コマンドリファレンス  | PPP の設定  | 233

[設定値及び初期値 ]
•divide
• [設定値 ] :
設定値 説明
on 分割する
off 分割しない
• [初期値 ] : on
[説明 ]
選択されている相手について [PPP, MP] に対して、 MP パケットの送信時にパケットを分割するか否かを設定する。
分割するとうまく接続できない相手に対してだけ  off にする。
分割しないように設定した場合、特に  TCP の転送効率に悪影響が出る可能性がある。
64 バイト以下のパケットは本コマンドの設定に関わらず分割されない。
[適用モデル ]
RTX5000, RTX3500, RTX1210
11.14 BACP 関連の設定
11.14.1 bacp-restart パラメータの設定
[書式 ]
ppp bacp  restart  time
no ppp bacp  restart  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : ミリ秒  (20..10000)
• [初期値 ] : 3000
[説明 ]
選択されている相手について [PPP,BACP] の configure-request 、terminate-request の再送時間を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3500, RTX1210
11.14.2 bacp-max-terminate パラメータの設定
[書式 ]
ppp bacp  maxterminate  count
no ppp bacp  maxterminate  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 2
[説明 ]
選択されている相手について [PPP,BACP] の terminate-request の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3500, RTX1210
11.14.3 bacp-max-configure パラメータの設定
[書式 ]
ppp bacp  maxconfigure  count
no ppp bacp  maxconfigure  [count ]234 | コマンドリファレンス  | PPP の設定

[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 10
[説明 ]
選択されている相手について [PPP, BACP] の configure-request の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3500, RTX1210
11.14.4 bacp-max-failure パラメータの設定
[書式 ]
ppp bacp  maxfailure  count
no ppp bacp  maxfailure  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 10
[説明 ]
選択されている相手について [PPP,BACP] の configure-nak の送信回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3500, RTX1210
11.15 BAP 関連の設定
11.15.1 bap-restart パラメータの設定
[書式 ]
ppp bap restart  time
no ppp bap restart  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : ミリ秒  (20..10000)
• [初期値 ] : 1000
[説明 ]
選択されている相手について [PPP,BAP] の configure-request 、terminate-request の再送時間を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3500, RTX1210
11.15.2 bap-max-retry パラメータの設定
[書式 ]
ppp bap maxretry  count
no ppp bap maxretry  [count ]
[設定値及び初期値 ]
•count
• [設定値 ] : 再送回数  (1..30)
• [初期値 ] : 30
[説明 ]
選択されている相手について [PPP,BAP] の最大再送回数を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3500, RTX1210コマンドリファレンス  | PPP の設定  | 235

11.16 PPPoE 関連の設定
11.16.1 PPPoE で使用する  LAN インタフェースの指定
[書式 ]
pppoe  use interface
no pppoe  use
[設定値及び初期値 ]
•interface
• [設定値 ] :
• LAN インタフェース名
• VLAN インタフェース名
• vRX シリーズ  および  RTX1300 では指定不可能
•タグ  LAN インタフェース名
• vRX シリーズ  および  RTX840、RTX830 では指定不可能
• [初期値 ] : -
[説明 ]
選択されている相手に対して、 PPPoE で使用するインタフェースを指定する。設定がない場合は、 PPPoE は使われ
ない。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.16.2 アクセスコンセントレータ名の設定
[書式 ]
pppoe  access  concentrator  name
no pppoe  access  concentrator
[設定値及び初期値 ]
•name
• [設定値 ] : アクセスコンセントレータの名前を表す文字列  (7bit US-ASCII)
• [初期値 ] : -
[説明 ]
選択されている相手について  PPPoE で接続するアクセスコンセントレータの名前を設定する。接続できるアクセス
コンセントレータが複数ある場合に、どのアクセスコンセントレータに接続するのかを指定するために使用する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.16.3 セッションの自動接続の設定
[書式 ]
pppoe  auto  connect  switch
no pppoe  auto  connect
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on 自動接続する
off 自動接続しない
• [初期値 ] : on
[説明 ]
選択されている相手に対して、 PPPoE のセッションを自動で接続するか否かを設定する。236 | コマンドリファレンス  | PPP の設定

[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.16.4 セッションの自動切断の設定
[書式 ]
pppoe  auto  disconnect  switch
no pppoe  auto  disconnect
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on 自動切断する
off 自動切断しない
• [初期値 ] : on
[説明 ]
選択されている相手に対して、 PPPoE のセッションを自動で切断するか否かを設定する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.16.5 PADI パケットの最大再送回数の設定
[書式 ]
pppoe  padi  maxretry  times
no pppoe  padi  maxretry
[設定値及び初期値 ]
•times
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 5
[説明 ]
PPPoE プロトコルにおける  PADI パケットの最大再送回数を設定する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.16.6 PADI パケットの再送時間の設定
[書式 ]
pppoe  padi  restart  time
no pppoe  padi  restart
[設定値及び初期値 ]
•time
• [設定値 ] : ミリ秒  (20..10000)
• [初期値 ] : 3000
[説明 ]
PPPoE プロトコルにおける  PADI パケットの再送時間を設定する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.16.7 PADR パケットの最大再送回数の設定
[書式 ]
pppoe  padr  maxretry  times
no pppoe  padr  maxretryコマンドリファレンス  | PPP の設定  | 237

[設定値及び初期値 ]
•times
• [設定値 ] : 回数  (1..10)
• [初期値 ] : 5
[説明 ]
PPPoE プロトコルにおける  PADR パケットの最大再送回数を設定する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.16.8 PADR パケットの再送時間の設定
[書式 ]
pppoe  padr  restart  time
no pppoe  padr  restart
[設定値及び初期値 ]
•time
• [設定値 ] : ミリ秒  (20..10000)
• [初期値 ] : 3000
[説明 ]
PPPoE プロトコルにおける  PADR パケットの再送時間を設定する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.16.9 PPPoE セッションの切断タイマの設定
[書式 ]
pppoe  disconnect  time  time
no pppoe  disconnect  time
[設定値及び初期値 ]
•time
• [設定値 ] :
設定値 説明
1..21474836 秒数
off タイマを設定しない
• [初期値 ] : off
[説明 ]
選択されている相手に対して、タイムアウトにより  PPPoE セッションを自動切断する時間を設定する。
[ノート ]
LCP と NCP パケットは監視対象外。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.16.10 サービス名の指定
[書式 ]
pppoe  service-name  name
no pppoe  service-name
[設定値及び初期値 ]
•name
• [設定値 ] : サービス名を表す文字列  (7bit US-ASCII 、255 文字以内  )
• [初期値 ] : -238 | コマンドリファレンス  | PPP の設定

[説明 ]
選択されている相手について  PPPoE で要求するサービス名を設定する。
接続できるアクセスコンセントレータが複数ある場合に、要求するサービスを提供することが可能なアクセスコン
セントレータを選択して接続するために使用する。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.16.11 TCP パケットの  MSS の制限の有無とサイズの指定
[書式 ]
pppoe  tcp mss limit  length
no pppoe  tcp mss limit
[設定値及び初期値 ]
•length
• [設定値 ] :
設定値 説明
1240..1452 データ長
auto MSS を MTU の値に応じて制限する
off MSS を制限しない
• [初期値 ] : auto
[説明 ]
PPPoE セッション上で  TCP パケットの  MSS(Maximum Segment Size) を制限するか否かを設定する。
[ノート ]
このコマンドと ip interface  tcp mss limit  コマンドの両方が有効な場合は、 MSS はどちらかより小さな方の値に制限
される。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
11.16.12 ルーター側には存在しない  PPPoE セッションを強制的に切断するか否かの設定
[書式 ]
pppoe  invalid-session  forced  close  sw
no pppoe  invalid-session  forced  close
[設定値及び初期値 ]
•sw
• [設定値 ] :
設定値 説明
on ルーター側には存在しない  PPPoE セッションを強制
的に切断する
off ルーター側には存在しない  PPPoE セッションを強制
的に切断しない
• [初期値 ] : on
[説明 ]
ルーター側には存在しない  PPPoE セッションを強制的に切断するか否かを設定します。
[適用モデル ]
vRX VMware ESXi 版, RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | PPP の設定  | 239

11.16.13 PPPoE フレームを中継するインターフェースの指定
[書式 ]
pppoe  pass-through  member  interface  interface  [interface ...]
no pppoe  pass-through  member  [...]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インターフェース名
• [初期値 ] : -
[説明 ]
PPPoE パススルー機能を使用するインターフェースを指定する。
指定したインターフェース間で  PPPoE フレームが中継される。
LAN インターフェース名には、物理  LAN インターフェースおよび  LAN 分割機能で使用するインターフェースを指
定できる。
[ノート ]
指定した  LAN インターフェースはプロミスキャスモードで動作する。
RTX830 は Rev.15.02.03 以降で使用可能。
RTX1210 は Rev.14.01.26 以降で使用可能。
RTX5000 、RTX3500 は Rev.14.00.26 以降で使用可能。
設定できる  interface  の数は下記の通り。
機種 最大設定可能数
RTX5000 10
RTX3510 10
RTX3500 10
RTX1300 8
RTX1220 10
RTX1210 10
RTX840 5
RTX830 5
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830240 | コマンドリファレンス  | PPP の設定

