# 第12章: DHCP の設定

> 元PDFページ: 241-258

---

第 12 章
DHCP の設定
本機は  DHCP(*1) 機能として、 DHCP サーバー機能、 DHCP リレーエージェント機能、 DHCP クライアント機能を実
装しています。
DHCP 機能の利用により、基本的なネットワーク環境の自動設定を実現します。
DHCP クライアント機能は  Windows 等の  OS に実装されており、これらと本機の  DHCP サーバー機能、 DHCP リレ
ーエージェント機能を組み合わせることにより  DHCP クライアントの基本的なネットワーク環境の自動設定を実現
します。
ルーターが  DHCP サーバーとして機能するか  DHCP リレーエージェントとして機能するか、どちらとしても機能さ
せないかは dhcp service  コマンドにより設定します。現在の設定は、 show status dhcp  コマンドにより知ることがで
きます。
DHCP サーバー機能は、 DHCP クライアントからのコンフィギュレーション要求を受けて  IP アドレスの割り当て
( リース  ) や、ネットマスク、 DNS サーバーの情報等を提供します。
割り当てる  IP アドレスの範囲とリース期間は dhcp scope  コマンドにより設定されたものが使用されます。
IP アドレスの範囲は複数の設定が可能であり、それぞれの範囲を  DHCP スコープ番号で管理します。 DHCP クライ
アントからの設定要求があると  DHCP サーバーは  DHCP スコープの中で未割り当ての  IP アドレスを自動的に通知
します。なお、特定の  DHCP クライアントに特定の  IP アドレスを固定的にリースする場合には、 dhcp scope  コマン
ドで定義したスコープ番号を用いて dhcp scope bind  コマンドで予約します。予約の解除は no dhcp scope bind  コマ
ンドで行います。 IP アドレスのリース期間には時間指定と無期限の両方が可能であり、これは dhcp scope  コマンド
の expire および  maxexpire キーワードのパラメータで指定します。
リース状況は show status dhcp  コマンドにより知ることができます。 DHCP クライアントに通知する  DNS サーバー
の IP アドレス情報は、 dns server  コマンドで設定されたものを使用します。
DHCP リレーエージェント機能は、ローカルセグメントの  DHCP クライアントからの要求を、予め設定されたリモ
ートのネットワークセグメントにある  DHCP サーバーへ転送します。リモートセグメントの  DHCP サーバーは
dhcp relay server  コマンドで設定します。 DHCP サーバーが複数ある場合には、 dhcp relay select  コマンドにより選
択方式を指定することができます。
また  DHCP クライアント機能により、 インタフェースの  IP アドレスやデフォルト経路情報などを外部の  DHCP サー
バーから受けることができます。ルーターを  DHCP クライアントとして機能させるかどうかは、 ip interface  address、
ip interface  secondary address 、ip pp remote address 、ip pp remote address pool  の各コマンドの設定値により決定さ
れます。設定されている内容は、 show status dhcpc  コマンドにより知ることができます。
(*1)Dynamic Host Configuration Protocol; RFC1541 , RFC2131
12.1 DHCP サーバー・リレーエージェント機能
12.1.1 DHCP の動作の設定
[書式 ]
dhcp  service  type
no dhcp  service  [type]
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
server DHCP サーバーとして機能させる
relay DHCP リレーエージェントとして機能させる
• [初期値 ] : -コマンドリファレンス  | DHCP の設定  | 241

[説明 ]
DHCP に関する機能を設定する。
DHCP リレーエージェント機能使用時には、 NAT 機能を使用することはできない。
[ノート ]
工場出荷状態および cold start  コマンド実行後の本コマンドの設定値については「 1.7 工場出荷状態またはデプロイ
後の初期状態における設定値」を参照してください。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.2 RFC2131 対応動作の設定
[書式 ]
dhcp  server  rfc2131  compliant  comp
dhcp  server  rfc2131  compliant  [except] function  [function ..]
no dhcp  server  rfc2131  compliant
[設定値及び初期値 ]
•comp
• [設定値 ] :
設定値 説明
on RFC2131 準拠
off RFC1541 準拠
• [初期値 ] : on
• except : 指定した機能以外が  RFC2131 対応となるキーワード
• [初期値 ] : -
•function
• [設定値 ] :
設定値 説明
broadcast-nak DHCPNAK をブロードキャストで送る
none-domain-null ドメイン名の最後に  NULL 文字を付加しない
remain-silent リース情報を持たないクライアントからの
DHCPREQUEST を無視する
reply-ack DHCPNAK の代わりに許容値を格納した  DHCPACK
を返す
use-clientid クライアントの識別に  Client-Identifier オプションを
優先する
• [初期値 ] : -
[説明 ]
DHCP サーバーの動作を指定する。 on の場合には  RFC2131 準拠となる。 off の場合には、 RFC1541 準拠の動作とな
る。
また  RFC1541 をベースとして  RFC2131 記述の個別機能のみを対応させる場合には以下のパラメータで指定する。
これらのパラメータはスペースで区切り複数指定できる。 except キーワードを指示すると、指定したパラメータ以
外の機能が  RFC2131 対応となる。
broadcast-nak同じサブネット上のクライアントに対しては
DHCPNAK はブロードキャストで送る。
DHCPREQUEST をクライアントが  INIT-REBOOT state
で送られてきたものに対しては、 giaddr 宛であれば  Bbit
を立てる。242 | コマンドリファレンス  | DHCP の設定

none-domain-null本ドメイン名の最後に  NULL 文字を付加しない。
RFC1541 ではドメイン名の最後に  NULL 文字を付加す
るかどうかは明確ではなかったが、 RFC2131 では禁止さ
れた。一方、 Windows NT/2000 の DHCP サーバーは
NULL 文字を付加している。そのため、 Windows 系の
OS での  DHCP クライアントは  NULL 文字があることを
期待している節があり、 NULL 文字がない場合には
winipcfg.exe での表示が乱れるなどの問題が起きる可能
性がある。
remain-silentクライアントから  DHCPREQUEST を受信した場合に、
そのクライアントのリース情報を持っていない場合に
は DHCPNAK を送らないようにする。
reply-ackクライアントから、リース期間などで許容できないオプ
ション値  ( リクエスト  IP アドレスは除く  ) を要求され
た場合でも、 DHCPNAK を返さずに許容値を格納した
DHCPACK を返す。
use-clientidクライアントの識別に  chaddr フィールドより  Client-
Identifier オプションを優先して使用する。
[ノート ]
工場出荷状態および cold start  コマンド実行後の本コマンドの設定値については「 1.7 工場出荷状態またはデプロイ
後の初期状態における設定値」を参照してください。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.3 リースする  IP アドレスの重複をチェックするか否かの設定
[書式 ]
dhcp  duplicate  check  check1  check2
no dhcp  duplicate  check
[設定値及び初期値 ]
•check1  : LAN 内を対象とするチェックの確認用待ち時間
• [設定値 ] :
設定値 説明
1..1000 ミリ秒
off LAN 内を対象とするチェックを行わない
• [初期値 ] : 100
•check2  : LAN 外 (DHCP リレーエージェント経由  ) を対象とするチェックの確認用待ち時間
• [設定値 ] :
設定値 説明
1..3000 ミリ秒
off LAN 外 (DHCP リレーエージェント経由  ) を対象と
するチェックを行わない
• [初期値 ] : 500
[説明 ]
DHCP サーバーとして機能する場合、 IP アドレスを  DHCP クライアントにリースする直前に、その  IP アドレスを使
っているホストが他にいないことをチェックするか否かを設定する。
[ノート ]
LAN 内のスコープに対しては  ARP を、DHCP リレーエージェント経由のスコープに対しては  PING を使ってチェッ
クする。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | DHCP の設定  | 243

12.1.4 DHCP スコープの定義
[書式 ]
dhcp  scope  scope_num  ip_address-ip_address /netmask  [except ex_ip  ...] [gateway gw_ip ] [expire time] [maxexpire time]
no dhcp  scope  scope_num  [ip_address-ip_address /netmask  [except ex_ip ...] [gateway gw_ip ] [expire time] [maxexpire
time]]
[設定値及び初期値 ]
•scope_num
• [設定値 ] : スコープ番号  (1..65535)
• [初期値 ] : -
•ip_address-ip_address
• [設定値 ] : 対象となるサブネットで割り当てる  IP アドレスの範囲
• [初期値 ] : -
•netmask
• [設定値 ] :
• xxx.xxx.xxx.xxx(xxx は十進数  )
• 0x に続く十六進数
•マスクビット数
• [初期値 ] : -
•ex_ip
• [設定値 ] : IP アドレス指定範囲の中で除外する  IP アドレス  ( 空白で区切って複数指定可能、 '-' を使用して範囲
指定も可能  )
• [初期値 ] : -
•gw_ip
• [設定値 ] : IP アドレス対象ネットワークのゲートウェイの  IP アドレス
• [初期値 ] : -
•time : 時間
• [設定値 ] :
• expire time : DHCP クライアントからリース期間要求がない場合のリース期間
• maxexpire time : DHCP クライアントからリース期間要求がある場合の許容最大リース期間
設定値 説明
1..2147483647分 ( RTX5000 / RTX3500 Rev.14.00.25 以前、 RTX1210
Rev.14.01.31 以前、および  RTX830 Rev.15.02.07 以前  )
1..21474836 分 ( 上記以外  )
xx:xx 時間  : 分
infinity 無期限リース
• [初期値 ] :
• expire time=72:00
• maxexpire time=72:00
[説明 ]
DHCP サーバーとして割り当てる  IP アドレスのスコープを設定する。
除外  IP アドレスは複数指定できる。リース期間としては無期限を指定できるほか、 DHCP クライアントから要求が
あった場合の許容最大リース期間を指定できる。
[ノート ]
同一ネットワークの DHCPスコープを複数設定できる  ( RTX5000 / RTX3500 Rev.14.00.25 以前、 RTX1210
Rev.14.01.27 以前、および  RTX830 Rev.15.02.02 以前を除く  ) 。
複数の  DHCP スコープで同一の  IP アドレスを含めることはできない。 IP アドレス範囲にネットワークアドレス、 ブ
ロードキャストアドレスを含む場合、割り当て可能アドレスから除外される。
vRX シリーズ  では、 netmask  の設定値を  /16 (255.255.0.0) から  /32 (255.255.255.255) の範囲に収めなければならない。244 | コマンドリファレンス  | DHCP の設定

DHCP リレーエージェントを経由しない  DHCP クライアントに対して  gateway キーワードによる設定パラメータが
省略されている場合にはルーター自身の  IP アドレスを通知する。
expire の設定値は  maxexpire の設定値以下でなければならない。
工場出荷状態および cold start  コマンド実行後の本コマンドの設定値については「 1.7 工場出荷状態またはデプロイ
後の初期状態における設定値」を参照してください。
RTX5000 / RTX3500 の Rev.14.00.21 以前、および、  RTX1210 の Rev.14.01.20 以前では、 dhcp scope  コマンドを実行
した場合に、同一のスコープ  ID を持つ以下のコマンドの設定が消去される。
•dhcp scope bind
•dhcp scope option
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.5 DHCP 予約アドレスの設定
[書式 ]
dhcp  scope  bind  scope_num  ip_address  [type] id
dhcp  scope  bind  scope_num  ip_address  mac_address
dhcp  scope  bind  scope_num  ip_address  ipcp
dhcp  scope  bind  scope_num  ip_address-ip_address  mac_address
no dhcp  scope  bind  scope_num  ip_address
no dhcp  scope  bind  scope_num  ip_address-ip_address
[設定値及び初期値 ]
•scope_num
• [設定値 ] : スコープ番号  (1..65535)
• [初期値 ] : -
•ip_address
• [設定値 ] :
設定値 説明
xxx.xxx.xxx.xxx (xxx は十進数  ) 予約する  IP アドレス
* 割り当てる  IP アドレスを指定しない
• [初期値 ] : -
•type : Client-Identifier オプションの  type フィールドを決定する
• [設定値 ] :
設定値 説明
text 0x00
ethernet 0x01
• [初期値 ] : -
•id
• [設定値 ] :
設定値 説明
type が ethernet の場合 MAC アドレス
type が text の場合 文字列
type が省略された場合 2 桁十六進数の列で先頭は  type フィールド
• [初期値 ] : -
•mac_address
• [設定値 ] :
• xx:xx:xx:xx:xx:xx(xx は十六進数  ) 予約  DHCP クライアントの  MAC アドレス
• xx:xx:xx:* のように下位  3 オクテットをアスタリスク  (*) にすることで、 OUI( ベンダー  ID) のみの指定とな
るコマンドリファレンス  | DHCP の設定  | 245

• [初期値 ] : -
• ipcp : IPCP でリモート側に与えることを示すキーワード
• [初期値 ] : -
[説明 ]
IP アドレスを割り当てる  DHCP クライアントを固定的に設定する。
IP アドレスを固定せずにクライアントだけを指定することもできる。この形式を削除する場合はクライアント識別
子を省略できない。
[ノート ]
IP アドレスは、 scope_num  パラメータで指定された  DHCP スコープ範囲内でなければならない。 1 つの  DHCP スコ
ープ内では、 1 つの  MAC アドレスに複数の  IP アドレスを設定することはできない。他の  DHCP クライアントにリ
ース中の  IP アドレスを予約設定した場合、リース終了後にその  IP アドレスの割り当てが行われる。
ipcp の指定は、同時に接続できる  B チャネルの数に限られる。また、 IPCP で与えるアドレスは  LAN 側のスコープ
から選択される。
コマンドの第  1 書式を使う場合は、あらかじめ  dhcp server rfc2131 compliant  on あるいは  use-clientid 機能を使用す
るよう設定されていなければならない。また  dhcp server rfc2131 compliant  off あるいは  use-clientid 機能が使用され
ないよう設定された時点で、コマンドの第  2 書式によるもの以外の予約は消去される。
コマンドの第  1 書式でのクライアント識別子は、クライアントがオプションで送ってくる値を設定する。 type パラ
メータを省略した場合には、 type フィールドの値も含めて入力する。 type パラメータにキーワードを指定する場合に
は type フィールド値は一意に決定されるので  Client-Identifier フィールドの値のみを入力する。
コマンドの第  2 書式による  MAC アドレスでの予約は、クライアントの識別に  DHCP パケットの  chaddr フィールド
を用いる。この形の予約機能は、 RT の設定が  dhcp server rfc2131 compliant  off あるいは  use-clientid 機能を使用しな
い設定になっているか、もしくは  DHCP クライアントが  DHCP パケット中に  Client-Identifier オプションを付けてこ
ない場合でないと動作しない。
クライアントが  Client-Identifier オプションを使う場合、コマンドの第  2 書式での予約は、 dhcp server rfc2131
compliant  on あるいは  use-clientid パラメータが指定された場合には無効になるため、新たに  Client-Identifier オプシ
ョンで送られる値で予約し直す必要がある。
コマンドの第 2書式で 1つの OUI(ベンダー ID)を複数設定することができる。 OUI(ベンダー ID)設定と MACアド
レス設定の両方がある場合、 MACアドレス設定を優先する。
OUI(ベンダー ID)設定は以下の機種・リビジョンでは指定不可能。
RTX5000 / RTX3500 Rev.14.00.25 以前
RTX1210 Rev.14.01.27 以前
RTX830 Rev.15.02.02 以前
RTX5000 / RTX3500 の Rev.14.00.21 以前、および、  RTX1210 の Rev.14.01.20 以前では、 dhcp scope  コマンドを実行
した場合に、同一のスコープ  ID を持つ以下のコマンドの設定が消去される。
•dhcp scope bind
•dhcp scope option
[設定例 ]
A. # dhcp scope bind 1 192.168.100.2 ethernet 00:a0:de:01:23:45
B. # dhcp scope bind 1 192.168.100.2 text client01
C. # dhcp scope bind 1 192.168.100.2 01 00 a0 de 01 23 45 01 01 01
D. # dhcp scope bind 1 192.168.100.2 00:a0:de:01:23:45
E. # dhcp scope bind 1 192.168.100.2-192.168.100.19 00:a0:de:*
1. dhcp server rfc2131 compliant  on あるいは  use-clientid 機能を使用する設定の場合
• A. B. C. の書式では、クライアントの識別に  Client-Identifier オプションを使用する。
• D. の書式では  DHCP パケットの  chaddr フィールドを使用する。ただし、 Client-Identifier オプションが存在する場
合、この設定は無視される。
DHCP サーバーは  chaddr フィールドの値より  Client-Identifier オプションの値の方が優先して使用される。
show status dhcp  コマンドを実行してクライアントの識別子を確認することで、クライアントが  Client-Identifier オプ
ションを使っているか否かを判別することも可能である。246 | コマンドリファレンス  | DHCP の設定

•リースしているクライアントとして  MAC アドレスが表示されていれば  Client-Identifier オプションは使用してい
ない
•リースしているクライアントとして十六進数の文字列、あるいは文字列が表示されていれば、 Client-Identifier オ
プションが使われている  Client-Identifier オプションを使うクライアントへの予約は、ここに表示される十六進数
の文字列あるいは文字列を使用する
2. dhcp server rfc2131 compliant  off あるいは  use-clientid 機能を使用しない場合
• A. B. C. の書式では指定できない。 Client-Identifier オプションは無視される。
• D. の書式では  DHCP パケットの  chaddr フィールドを使用する。
なお、クライアントとの相互動作に関して以下の留意点がある。
•個々の機能を単独で用いるとクライアント側で思わぬ動作を招く可能性があるため、 dhcp server rfc2131
compliant  on あるいは  dhcp server rfc2131 compliant  off で使用することを推奨する。
•ルーターの再起動やスコープの再設定によりリース情報が消去されている場合、アドレスの延長要求をした時や
リース期間内のクライアントを再起動した時にクライアントが使用する  IP アドレスは変わることがある。
これを防ぐためには  dhcp server rfc2131 compliant  on ( あるいは  remain-silent 機能を有効にする  ) 設定がある。
この設定にすると、ヤマハルーターがリース情報を持たないクライアントからの  DHCPREQUEST に対して
DHCPNAK を返さず無視するようになる。
この結果、リース期限満了時にクライアントが出す  DHCPDISCOVER に Requested IP Address オプションが含ま
れていれば、そのクライアントには引き続き同じ  IP アドレスをリースすることができる。
E.の書式では、 OUI(ベンダー ID)のみ指定し、その OUI(ベンダー ID)を持つ機器にのみ IPアドレスを割り当てるこ
とができる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.6 DHCP アドレス割り当て動作の設定
[書式 ]
dhcp  scope  lease  type scope_num  type [fallback= fallback_scope_num ]
no dhcp  scope  lease  type scope_num  [type ...]
[設定値及び初期値 ]
•scope_num, fallback_scope_num
• [設定値 ] : スコープ番号  (1-65535)
• [初期値 ] : -
•type : 割り当ての動作
• [設定値 ] :
設定値 説明
bind-priority 予約情報を優先して割り当てる
bind-only 予約情報だけに制限して割り当てる
• [初期値 ] : bind-priority
[説明 ]
scope_num  で指定した  DHCP スコープにおける、アドレスの割り当て方法を制御する。
type に bind-priority を指定した場合には、 dhcp scope bind  コマンドで予約されたクライアントには予約どおりの  IP
アドレスを、予約されていないクライアントには他のクライアントに予約されていない空きアドレスがスコープ内
にある限りそれを割り当てる。
type に bind-priority を指定した場合には、 fallback オプションは指定できない。
type に bind-only を指定した場合は、 fallback オプションでフォールバックスコープを指定しているかどうかによって
動作が変わる。
fallback オプションの指定が無い場合、 dhcp scope bind  コマンドで予約されているクライアントにのみ  IP アドレス
を割り当て、予約されていないクライアントにはたとえスコープに空きがあっても  IP アドレスを割り当てない。
type に bind-only を指定し、同時に  fallback オプションでフォールバックスコープを指定している場合には、以下の
ような動作になる。コマンドリファレンス  | DHCP の設定  | 247

1.クライアントが、スコープで  IP アドレスを予約されている時には、予約どおりの  IP アドレスを割り当てる。
2.クライアントが、スコープでは  IP アドレスが予約されていないが、フォールバックスコープでは予約されている
時には、フォールバックスコープでの予約どおりの  IP アドレスを割り当てる。
3.クライアントが、スコープ、フォールバックスコープのいずれでも  IP アドレスを予約されていない時には、フォ
ールバックスコープに対する dhcp scope lease type  コマンドの設定によって動作が変わる。
a.フォールバックスコープに対する dhcp scope lease type  コマンドの設定が  bind-priority になっている時には、
クライアントにはフォールバックスコープに空きアドレスがある限りそれを割り当てる。
b.フォールバックスコープに対する dhcp scope lease type  コマンドの設定が  bind-only になっている時には、 クラ
イアントには  IP アドレスは割り当てられない。
いずれの場合も、リース期間は各  DHCP スコープの定義に従う。
リース期間は  DHCP スコープの定義に従う。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.7 DHCP 割り当て情報を元にした予約設定の生成
[書式 ]
dhcp  convert  lease  to bind  scope_n  [except] [ idx [...]]
[設定値及び初期値 ]
•scope_n
• [設定値 ] : スコープ番号  (1-65535)
• [初期値 ] : -
•idx
• [設定値 ] :
設定値 説明
番号 show status dhcp summary  コマンドで表示されるイ
ンデックス番号、最大  100 個
all 割り当て中の情報全てを対象とする
省略 省略時は  all
• [初期値 ] : -
[説明 ]
現在の割り当て情報を元に予約設定を作成する。 except キーワードを指示すると、指定した番号以外の情報が予約
設定に反映される。
[ノート ]
以下の変換規則で  IP アドレス割り当て情報が予約設定に変換される。
IP アドレス割り当て情報のクライア
ント識別種別  (show status dhcp で表
示される名称  )クライアント識別情報例 予約設定情報例
クライアントイーサネットアドレス 00:a0:de:01:02:03ethernet 00:a0:de:01:02:03 ※1
00:a0:de:01:02:03 ※2
クライアント  ID(01) 00 a0 de 01 02 03 ethernet 00:a0:de:01:02:03
(01) 00 a0 de 01 02 03 04 01 00 a0 de 01 02 03 04
(01) 31 32 33 00 31 32 33
※1：rfc2131 compliant on あるいは  use-clientid ありの場合、このような  IP アドレス割り当て情報の表示は  ARP チ
ェックの結果である可能性が高く、通常の割り当て時にはクライアント  ID オプションが使われるため、この形
式で予約設定をする。  ただし、 MAC アドレスと異なるクライアント  ID を使うホストが存在する場合はこの自動
変換による予約は有効に機能しないため、そのようなホストに対する予約設定は別途、手動で行う必要がある
※2：rfc2131 compliant off あるいは  use-clientid なしの場合、 chaddr フィールドを使用する248 | コマンドリファレンス  | DHCP の設定

コマンド実行時点での割り当て情報を元に予約設定を作成する。サマリ表示からこの変換コマンドの実行までに時
間が経過した場合には、本コマンド実行後に意図したペアの予約が作成されていることを show config  で確認するべ
きである
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.8 DHCP オプションの設定
[書式 ]
dhcp  scope  option  scope_num  option =value  [option =value ...]
no dhcp  scope  option  scope_num  [...]
[設定値及び初期値 ]
•scope_num
• [設定値 ] : スコープ番号  (1..65535)
• [初期値 ] : -
•option
• [設定値 ] :
•オプション番号
• 1..49,62..254
•ニーモニック
•主なニーモニック
router 3
dns 6
hostname 12
domain 15
wins_server 44
• [初期値 ] : -
•value  : オプション値
• [設定値 ] :
•値としては以下の種類があり、どれが使えるかはオプション番号で決まる。例え
ば、 'router','dns','wins_server' は IP アドレスの配列であり、 'hostname','domain' は文字列である。
1 オクテット整数 0..255
2 オクテット整数 0..65535
2 オクテット整数の配列 2 オクテット整数をコンマ  (,) で並べたもの
4 オクテット整数 0..2147483647
IP アドレス IP アドレス
IP アドレスの配列 IP アドレスをコンマ  (,) で並べたもの
文字列 文字列
スイッチ "on","off","1","0" のいずれか
バイナリ 2 桁十六進数をコンマ  (,) で並べたもの
• [初期値 ] : -
[説明 ]
スコープに対して送信する  DHCP オプションを設定する。 dns server  コマンドや wins server  コマンドなどでも暗黙
のうちに  DHCP オプションを送信していたが、それを明示的に指定できる。また、暗黙の  DHCP オプションではス
コープでオプションの値を変更することはできないが、このコマンドを使えばそれも可能になる。
[ノート ]
RTX5000 / RTX3500 の Rev.14.00.21 以前、および、  RTX1210 の Rev.14.01.20 以前では、 dhcp scope  コマンドを実行
した場合に、同一のスコープ  ID を持つ以下のコマンドの設定が消去される。コマンドリファレンス  | DHCP の設定  | 249

•dhcp scope bind
•dhcp scope option
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.9 DHCP リース情報の手動追加
[書式 ]
dhcp  manual  lease  ip_address  [type] id
dhcp  manual  lease  ip_address  mac_address
dhcp  manual  lease  ip_address  ipcp
[設定値及び初期値 ]
•ip_address
• [設定値 ] : リースする  IP アドレス
• [初期値 ] : -
•type : Client-Identifier オプションの  type フィールドを決定する
• [設定値 ] :
設定値 説明
text 0x00
ethernet 0x01
• [初期値 ] : -
•id
• [設定値 ] :
設定値 説明
type が text の場合 文字列
type が ethernet の場合 MAC アドレス
type が省略された場合 2 桁十六進数の列で先頭は type フィールド
• [初期値 ] : -
•mac_address
• [設定値 ] : XX:XX:XX:XX:XX:XX(XX は十六進数  )DHCP クライアントの  MAC アドレス
• [初期値 ] : -
• ipcp : IPCP でリモート側に与えたものとするキーワード
• [初期値 ] : -
[説明 ]
手動で、特定  IP アドレスのリース情報を追加する。
[ノート ]
本コマンドは自動で行われる  DHCP のアドレス配布に影響を与えるため、意図して特定の  IP アドレスのリース情報
を追加したい場合を除いて、使用するべきではない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.10 DHCP リース情報の手動削除
[書式 ]
dhcp  manual  release  ip_address
[設定値及び初期値 ]
•ip_address
• [設定値 ] : 解放する  IP アドレス
• [初期値 ] : -250 | コマンドリファレンス  | DHCP の設定

[説明 ]
手動で、特定  IP アドレスのリース情報を削除する。
[ノート ]
本コマンドは自動で行われる  DHCP のアドレス配布に影響を与えるため、意図して特定の  IP アドレスのリース情報
を削除したい場合を除いて、使用するべきではない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.11 DHCP サーバーの指定の設定
[書式 ]
dhcp  relay  server  host1  [host2  [host3  [host4 ]]]
no dhcp  relay  server
[設定値及び初期値 ]
•host1..host4
• [設定値 ] : DHCP サーバーの  IP アドレス
• [初期値 ] : -
[説明 ]
DHCPBOOTREQUEST パケットを中継するサーバーを最大  4 つまで設定する。
サーバーが複数指定された場合は、 BOOTREQUEST パケットを複写してすべてのサーバーに中継するか、あるいは
1 つだけサーバーを選択して中継するかは dhcp relay select  コマンドの設定で決定される。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.12 DHCP リレーエージェント機能で使用する始点ポート番号の設定
[書式 ]
dhcp  relay  srcport  port
no dhcp  relay  srcport  [port]
[設定値及び初期値 ]
•port
• [設定値 ] : ポート番号  (1..65535)
• [初期値 ] : 68
[説明 ]
DHCP リレーエージェント機能で使用する始点ポート番号を設定する。
[ノート ]
RTX5000 、RTX3500 は Rev.14.00.29 以降、 RTX1210 Rev.14.01.36 以降、 RTX830 Rev.15.02.15 以降で使用可能。
[適用モデル ]
vRX VMware ESXi 版, vRX さくらのクラウド版 , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210,
RTX840, RTX830
12.1.13 DHCP サーバーの選択方法の設定
[書式 ]
dhcp  relay  select  type
no dhcp  relay  select  [type]
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
hash Hash 関数を利用して一つだけサーバーを選択するコマンドリファレンス  | DHCP の設定  | 251

設定値 説明
all すべてのサーバーを選択する
• [初期値 ] : hash
[説明 ]
dhcp relay server  コマンドで設定された複数のサーバーの取り扱いを設定する。
hash が指定された場合は、 Hash 関数を利用して一つだけサーバーが選択されてパケットが中継される。この  Hash
関数は、 DHCP メッセージの  chaddr フィールドを引数とするので、同一の  DHCP クライアントに対しては常に同じ
サーバーが選択されるはずである。 all が指定された場合は、パケットはすべてのサーバーに対し複写中継される。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.14 DHCP BOOTREQUEST パケットの中継基準の設定
[書式 ]
dhcp  relay  threshold  time
no dhcp  relay  threshold  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : 秒数  (0..65535)
• [初期値 ] : 0
[説明 ]
DHCP BOOTREQUEST パケットの  secs フィールドとこのコマンドによる秒数を比較し、 設定値より小さな  secs フィ
ールドを持つ  DHCP BOOTREQUEST パケットはサーバーに中継しないようにする。
これにより、同一  LAN 上に別の  DHCP サーバーがあるにも関わらず遠隔地の  DHCP サーバーにパケットを中継し
てしまうのを避けることができる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.1.15 インターフェース毎の DHCPの動作の設定
[書式 ]
ip interface  dhcp  service  type [host1  [host2  [host3  [host4 ]]]]
no ip interface  dhcp  service
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、ブリッジインタフェース名
• [初期値 ] : -
•type
• [設定値 ] :
設定値 説明
off DHCP サーバーとしても  DHCP リレーエージェント
としても機能しない
server DHCP サーバーとして機能させる
relay DHCP リレーエージェントとして機能させる
• [初期値 ] : -
•host1..host4
• [設定値 ] : DHCP サーバーの  IP アドレス
• [初期値 ] : -252 | コマンドリファレンス  | DHCP の設定

[説明 ]
インターフェース毎に  DHCP の動作を設定する。
DHCP サーバーを設定した場合には、ネットワークアドレスが合致する  DHCP スコープから  IP アドレスを 1つ割り
当てる。
DHCP リレーエージェントを設定した場合には、 HOST を設定する必要があり、この  HOST へ DHCP DISCOVER パ
ケットおよび  DHCP REQUEST パケットを転送する。
off に設定した場合には、 DHCP サーバーとしても  DHCP リレーエージェントとしても動作しない。 DHCP パケット
は破棄されます。
本設定が無い場合は、 dhcp service コマンドの設定に従う。 dhcp service コマンドの設定と本設定の両方がある場合に
は、本設定が優先される。
[ノート ]
ブリッジインタフェースは、 vRX Amazon EC2 版 では指定不可能。
RTX5000 、RTX3500 は Rev.14.00.18 以降で使用可能。
RTX1210 は Rev.14.01.09 以降で使用可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.2 DHCP クライアント機能
12.2.1 DHCP クライアントのホスト名の設定
[書式 ]
dhcp  client  hostname  interface  primary host
dhcp  client  hostname  interface  secondary host
dhcp  client  hostname  pp peer_num  host
dhcp  client  hostname  pool pool_num  host
no dhcp  client  hostname  interface  primary [ host]
no dhcp  client  hostname  interface  secondary [ host]
no dhcp  client  hostname  pp peer_num  [host]
no dhcp  client  hostname  pool pool_num  [host]
[設定値及び初期値 ]
•interface
• [設定値 ] :
• LAN インタフェース名
• WAN インタフェース名
• vRX シリーズ、 RTX5000 、および  RTX3500 では指定不可能
•ブリッジインタフェース名
• [初期値 ] : -
•peer_num
• [設定値 ] :
•相手先情報番号
• anonymous
• [初期値 ] : -
•pool_num
• [設定値 ] : ip pp remote address pool dhcpc  コマンドで取得する  IP アドレスの番号。例えば、 ip pp remote
address pool dhcpc  コマンドで  IP アドレスを  2 個取得できる機種で、 pool_num  に "1" または  "2" を設定するこ
とで、それぞれのクライアント  ID オプションに任意の  ID を付けることができる。 (1..ip pp remote address
pool dhcpc  コマンドで取得できる  IP アドレスの最大数  )
• [初期値 ] : -
•host
• [設定値 ] : DHCP クライアントのホスト名
• [初期値 ] : -コマンドリファレンス  | DHCP の設定  | 253

[説明 ]
DHCP クライアントのホスト名を設定する。
[ノート ]
ブリッジインタフェースは、 vRX Amazon EC2 版、および  RTX5000 / RTX3500 Rev.14.00.11 以前では指定不可能。
WAN インタフェースを設定した時には、 secondary は指定できない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•peer_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 150
•pool_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 101 - - - -
YSL-VPN-EX2 301 501 701 901 1101
YSL-VPN-EX3 1541 2041 2541 3041 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.2.2 要求する  IP アドレスリース期間の設定
[書式 ]
ip interface  dhcp  lease  time  time
no ip interface  dhcp  lease  time  [time]
[設定値及び初期値 ]
•interface
• [設定値 ] :
• LAN インタフェース名
• WAN インタフェース名
• vRX シリーズ、 RTX5000 、および  RTX3500 では指定不可能
•ブリッジインタフェース名
• [初期値 ] : -
•time
• [設定値 ] : 分数  (1..21474836)
• [初期値 ] : -
[説明 ]
DHCP クライアントが要求する  IP アドレスのリース期間を設定する。
[ノート ]
リース期間の要求が受け入れられなかった場合、要求しなかった場合は、 DHCP サーバーからのリース期間を利用
する。
ブリッジインタフェースは、 vRX Amazon EC2 版、および  RTX5000 / RTX3500 Rev.14.00.11 以前では指定不可能。254 | コマンドリファレンス  | DHCP の設定

[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.2.3 IP アドレス取得要求の再送回数と間隔の設定
[書式 ]
ip interface  dhcp  retry  retry  interval
no ip interface  dhcp  retry  [retry  interval ]
[設定値及び初期値 ]
•interface
• [設定値 ] :
• LAN インタフェース名
• WAN インタフェース名
• vRX シリーズ、 RTX5000 、および  RTX3500 では指定不可能
•ブリッジインタフェース名
• [初期値 ] : -
•retry
• [設定値 ] :
設定値 説明
1..100 回数
infinity 無制限
• [初期値 ] : infinity
•interval
• [設定値 ] : 秒数  (1..100)
• [初期値 ] : 5
[説明 ]
IP アドレスの取得に失敗したときにリトライする回数とその間隔を設定する。
[ノート ]
ブリッジインタフェースは、 vRX Amazon EC2 版、および  RTX5000 / RTX3500 Rev.14.00.11 以前では指定不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.2.4 DHCP クライアント  ID オプションの設定
[書式 ]
dhcp  client  client-identifier  interface  primary [type type] id
dhcp  client  client-identifier  interface  secondary [type type] id
dhcp  client  client-identifier  pp peer_num  [type type] id
dhcp  client  client-identifier  pool pool_num  [type type] id
no dhcp  client  client-identifier  interface  primary
no dhcp  client  client-identifier  interface  secondary
no dhcp  client  client-identifier  pp peer_num
no dhcp  client  client-identifier  pool pool_num
[設定値及び初期値 ]
•interface
• [設定値 ] :
• LAN インタフェース名
• WAN インタフェース名
• vRX シリーズ、 RTX5000 、および  RTX3500 では指定不可能
•ブリッジインタフェース名コマンドリファレンス  | DHCP の設定  | 255

• [初期値 ] : -
• type : ID オプションの  type フィールドの値を設定することを示すキーワード
• [初期値 ] : -
•type
• [設定値 ] : ID オプションの  type フィールドの値
• [初期値 ] : 1
•id
• [設定値 ] :
• ASCII 文字列で表した  ID
• 2 桁の十六進数列で表した  ID
• [初期値 ] : -
•peer_num
• [設定値 ] :
•相手先情報番号
• anonymous
• [初期値 ] : -
•pool_num
• [設定値 ] : ip pp remote address pool dhcpc  コマンドで取得する  IP アドレスの番号。例えば、 ip pp remote
address pool dhcpc  コマンドで  IP アドレスを  2 個取得できる機種で、 pool_num  に "1" または  "2" を設定するこ
とで、それぞれのクライアント  ID オプションに任意の  ID を付けることができる。 (1..ip pp remote address
pool dhcpc  コマンドで取得できる  IP アドレスの最大数  )
• [初期値 ] : -
[説明 ]
DHCP クライアント  ID オプションの  type フィールドと  ID を設定する。
[ノート ]
ブリッジインタフェースは、 vRX Amazon EC2 版、および  RTX5000 / RTX3500 Rev.14.00.11 以前では指定不可能。
WAN インタフェースを設定した時には、 secondary は指定できない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•peer_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 150
•pool_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 101 - - - -
YSL-VPN-EX2 301 501 701 901 1101
YSL-VPN-EX3 1541 2041 2541 3041 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830256 | コマンドリファレンス  | DHCP の設定

12.2.5 DHCP クライアントが  DHCP サーバーへ送るメッセージ中に格納するオプションの設定
[書式 ]
dhcp  client  option  interface  primary option =value
dhcp  client  option  interface  secondary option =value
dhcp  client  option  pp peer_num  option =value
dhcp  client  option  pool pool_num  option =value
no dhcp  client  option  interface  primary [ option =value ]
no dhcp  client  option  interface  secondary [ option =value ]
no dhcp  client  option  pp peer_num  [option =value ]
no dhcp  client  option  pool pool_num  [option =value ]
[設定値及び初期値 ]
•interface
• [設定値 ] :
• LAN インタフェース名
• WAN インタフェース名
• vRX シリーズ、 RTX5000 、および  RTX3500 では指定不可能
•ブリッジインタフェース名
• [初期値 ] : -
•option
• [設定値 ] : オプション番号  ( 十進数  )
• [初期値 ] : -
•value
• [設定値 ] : 格納するオプション値  ( 十六進数、 "," で区切って複数指定可能  ) なおオプション長情報は入力の必
要はない
• [初期値 ] : -
•peer_num
• [設定値 ] :
•相手先情報番号
• anonymous
• [初期値 ] : -
•pool_num
• [設定値 ] : ip pp remote address pool dhcpc  コマンドで取得する  IP アドレスの番号。例えば、 ip pp remote
address pool dhcpc  コマンドで  IP アドレスを  2 個取得できる機種で、 pool_num  に "1" または  "2" を設定するこ
とで、それぞれのクライアント  ID オプションに任意の  ID を付けることができる。 (1..ip pp remote address
pool dhcpc  コマンドで取得できる  IP アドレスの最大数  )
• [初期値 ] : -
[説明 ]
DHCP クライアントが  DHCP サーバーへ送るメッセージ中に格納するオプションを設定する。
[ノート ]
このコマンドはサーバーとの相互接続に必要な場合にのみ設定する。
得られたオプション値は内部では利用されない。
ブリッジインタフェースは、 vRX Amazon EC2 版、および  RTX5000 / RTX3500 Rev.14.00.11 以前では指定不可能。
WAN インタフェースを設定した時には、 secondary は指定できない。
[設定例 ]
1.LAN2 プライマリアドレスを  DHCP サーバーから得る場合に特定アドレス  (192.168.0.128) を要求する。
# dhcp client option lan2 primary 50=c0,a8,00,80
# ip lan2 address dhcp
( 注：ただし、この場合でも要求アドレスがサーバーから与えられるか否かはサーバー次第である。 )
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。コマンドリファレンス  | DHCP の設定  | 257

•peer_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 150
•pool_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 101 - - - -
YSL-VPN-EX2 301 501 701 901 1101
YSL-VPN-EX3 1541 2041 2541 3041 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
12.2.6 リンクダウンした時に情報を解放するか否かの設定
[書式 ]
dhcp  client  release  linkdown  switch  [time]
no dhcp  client  release  linkdown  [switch  [time]]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on インタフェースのリンクダウンが time秒間継続する
と、取得していた情報を解放する
off インタフェースがリンクダウンしても情報は保持す
る
• [初期値 ] : off
•time
• [設定値 ] : 秒数  (0..259200)
• [初期値 ] : 3
[説明 ]
DHCPクライアントとして DHCPサーバーから IPアドレスを得ているインタフェースがリンクダウンした時に、
DHCPサーバーから得ていた情報を解放するか否かを設定する。
リンクダウンするとタイマーが働き、 timeの秒数だけリンクダウン状態が継続すると情報を解放する。 timeが設定
されていない場合には timeは3秒となる。
情報が解放されると、次にリンクアップした時に情報の取得を試みる。
[ノート ]
タイマーの値を長く設定すると、不安定なリンク状態の影響を避けることができる。
本コマンドの設定は、コマンド実行後に発生したリンクダウン以降で有効になる。
タイマーの満了前にリンクアップした場合にはタイマーはクリアされ、情報を解放しない。
タイマーの満了前に情報のリース期間が満了した場合には、タイマーはクリアされ、情報は解放される。
以下のコマンド実行時には、動作中のタイマーはクリアされる。
ip interface  address , ip pp remote address , ip pp remote address pool , dhcp client release linkdown
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830258 | コマンドリファレンス  | DHCP の設定

