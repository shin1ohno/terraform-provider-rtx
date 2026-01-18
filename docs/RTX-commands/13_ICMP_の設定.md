# 第13章: ICMP の設定

> 元PDFページ: 259-270

---

第 13 章
ICMP の設定
13.1 IPv4 の設定
13.1.1 ICMP Echo Reply を送信するか否かの設定
[書式 ]
ip icmp  echo-reply  send  send
no ip icmp  echo-reply  send  [send]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
[説明 ]
ICMP Echo を受信した場合に、 ICMP Echo Reply を返すか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.1.2 ICMP Echo Reply をリンクダウン時に送信するか否かの設定
[書式 ]
ip icmp  echo-reply  send-only-linkup  send
no ip icmp  echo-reply  send-only-linkup  [send]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on リンクアップしている時だけ  ICMP Echo Reply を返
す
off リンクの状態に関わらず  ICMP Echo Reply を返す
• [初期値 ] : off
[説明 ]
リンクダウンしているインタフェースに付与された  IP アドレスを終点  IP アドレスとする  ICMP Echo を受信した時
に、それに対して  ICMP Echo Reply を返すかどうかを設定する。 on に設定した時には、リンクアップしている時だ
け ICMP Echo を返すので、リンクの状態を  ping で調べることができるようになる。 off に設定した場合には、リン
クの状態に関わらず  ICMP Echo を返す。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.1.3 ICMP Mask Reply を送信するか否かの設定
[書式 ]
ip icmp  mask-reply  send  send
no ip icmp  mask-reply  send  [send]コマンドリファレンス  | ICMP の設定  | 259

[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
[説明 ]
ICMP Mask Request を受信した場合に、 ICMP Mask Reply を返すか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.1.4 ICMP Parameter Problem を送信するか否かの設定
[書式 ]
ip icmp  parameter-problem  send  send
no ip icmp  parameter-problem  send  [send]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : off
[説明 ]
受信した  IP パケットの  IP オプションにエラーを検出した場合に、 ICMP Parameter Problem を送信するか否かを設定
する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.1.5 ICMP Redirect を送信するか否かの設定
[書式 ]
ip icmp  redirect  send  send
no ip icmp  redirect  send  [send]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
[説明 ]
他のゲートウェイ宛の  IP パケットを受信して、そのパケットを適切なゲートウェイに回送した場合に、同時にパケ
ットの送信元に対して  ICMP Redirect を送信するか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830260 | コマンドリファレンス  | ICMP の設定

13.1.6 ICMP Redirect 受信時の処理の設定
[書式 ]
ip icmp  redirect  receive  action
no ip icmp  redirect  receive  [action ]
[設定値及び初期値 ]
•action
• [設定値 ] :
設定値 説明
on 処理する
off 無視する
• [初期値 ] : off
[説明 ]
ICMP Redirect を受信した場合に、それを処理して自分の経路テーブルに反映させるか、あるいは無視するかを設定
する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.1.7 ICMP Time Exceeded を送信するか否かの設定
[書式 ]
ip icmp  time-exceeded  send  send [rebound =sw]
no ip icmp  time-exceeded  send  [send rebound =sw]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
•sw
• [設定値 ] :
設定値 説明
on 受信インターフェースから送信する
off 経路に従って送信する
• [初期値 ] : off
[説明 ]
受信した  IP パケットの  TTL が 0 になってしまったため、そのパケットを破棄した場合に、同時にパケットの送信元
に対して  ICMP Time Exceeded を送信するか否かを設定する。
rebound オプションを  on に設定した場合には、経路に関係なく元となるパケットを受信したインターフェースから
送信する。
[ノート ]
rebound オプションは、 RTX5000 / RTX3500 Rev.14.00.11 以前では指定不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | ICMP の設定  | 261

13.1.8 ICMP Timestamp Reply を送信するか否かの設定
[書式 ]
ip icmp  timestamp-reply  send  send
no ip icmp  timestamp-reply  send  [send]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
[説明 ]
ICMP Timestamp を受信した場合に、 ICMP Timestamp Reply を返すか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.1.9 ICMP Destination Unreachable を送信するか否かの設定
[書式 ]
ip icmp  unreachable  send  send [rebound =sw]
no ip icmp  unreachable  send  [send rebound =sw]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
•sw
• [設定値 ] :
設定値 説明
on 受信インターフェースから送信する
off 経路に従って送信する
• [初期値 ] : off
[説明 ]
経路テーブルに宛先が見つからない場合や、 あるいは  ARP が解決できなくて  IP パケットを破棄することになった場
合に、同時にパケットの送信元に対して  ICMP Destination Unreachable を送信するか否かを設定する。
rebound オプションを  on に設定した場合には、経路に関係なく元となるパケットを受信したインターフェースから
送信する。
[ノート ]
rebound オプションは、 RTX5000 / RTX3500 Rev.14.00.11 以前では指定不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830262 | コマンドリファレンス  | ICMP の設定

13.1.10 IPsec で復号したパケットに対して  ICMP エラーを送るか否かの設定
[書式 ]
ip icmp  error-decrypted-ipsec  send  switch
no ip icmp  error-decrypted-ipsec  send  [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on IPsec で復号したパケットに対して  ICMP エラーを送
る
off IPsec で復号したパケットに対して  ICMP エラーを送
らない
• [初期値 ] : on
[説明 ]
IPsec で復号したパケットに対して  ICMP エラーを送るか否か設定する。
[ノート ]
ICMP エラーには復号したパケットの先頭部分が含まれるため、 ICMP エラーが送信元に返送される時にも  IPsec で
処理されないようになっていると、 本来  IPsec で保護したい通信が保護されずにネットワークに流れてしまう可能性
がある。特に、 フィルタ型ルーティングでプロトコルによって  IPsec で処理するかどうか切替えている場合には注意
が必要となる。
ICMP エラーを送らないように設定すると、 traceroute に対して反応がなくなるなどの現象になる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.1.11 受信した  ICMP のログを記録するか否かの設定
[書式 ]
ip icmp  log log
no ip icmp  log [log]
[設定値及び初期値 ]
•log
• [設定値 ] :
設定値 説明
on 記録する
off 記録しない
• [初期値 ] : off
[説明 ]
受信した  ICMP エラーを  DEBUG レベルのログに記録するか否かを設定する。 Echo Request や Echo Reply のログは
記録しない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.1.12 ステルス機能の設定
[書式 ]
ip stealth  all
ip stealth  interface  [interface ...]
no ip stealth  [...]
[設定値及び初期値 ]
• all : すべての論理インタフェースで受信したパケットに対してステルス動作を行うコマンドリファレンス  | ICMP の設定  | 263

• [初期値 ] : -
•interface
• [設定値 ] : 指定した論理インタフェースで受信したパケットに対してステルス動作を行う
• [初期値 ] : -
[説明 ]
このコマンドを設定すると、指定されたインタフェースからルーター宛に来たパケットが原因で発生する  ICMP お
よび  TCP リセットを返さないようになる。
ルーターがサポートしていないプロトコルや  IPv6 ヘッダ、あるいはオープンしていない  TCP/UDP ポートに対して
指定されたインタフェースからパケットを受信した時に、通常であれば  ICMP unreachable や TCP リセットを返送す
る。しかし、このコマンドを設定しておくとそれを禁止することができ、ポートスキャナーなどによる攻撃を受け
た時にルーターの存在を隠すことができる。
[ノート ]
このコマンドで指定したインタフェースで受信した  PING にも答えなくなるので注意が必要である。
ルーター宛ではないパケットが原因で発生する  ICMP はこのコマンドでは制御できない。それらを送信しないよう
にするには、  ip icmp  * コマンド群を用いる必要がある。
ブリッジインタフェースは  RTX5000 / RTX3500 Rev.14.00.12 以降のファームウェア、および、 Rev.14.01 系以降のす
べてのファームウェアで指定可能。
WAN インタフェースは  Rev.14.01 系以降のファームウェアで指定可能。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.1.13 ARP による  MTU 探索を行うか否かの設定
[書式 ]
ip interface  arp mtu discovery  sw [minimum= min_mtu ]
no ip interface  arp mtu discovery  [sw [minimum= min_mtu ]]
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名
• [初期値 ] : -
•sw
• [設定値 ] :
設定値 説明
on ARP による  MTU 探索を行う
off ARP による  MTU 探索を行わない
• [初期値 ] : on
•min_mtu
• [設定値 ] : 探索範囲の最低  MTU
• [初期値 ] : 4000
[説明 ]
ARP による  MTU 探索を行うか否かを設定します。
指定したインタフェースで、  lan type  コマンドおよび ip mtu  コマンドによりジャンボフレームが利用できる状況に
ある時にこのコマンドが  on と設定されていると、 ARP 解決できた相手に対して大きなサイズの  ARP を繰り返し送
ることで相手の  MTU を探索します。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300
13.1.14 切り詰められたパケットに対して、 ICMP Destination Unreachable を送信するか否かの設定
[書式 ]
ip icmp  unreachable-for-truncated  send  send264 | コマンドリファレンス  | ICMP の設定

no ip icmp  unreachable-for-truncated  send  [send]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
[説明 ]
LAN インタフェースで受信したが、そのインタフェースの  MTU を越える長さだったために切り詰められたパケッ
トに対して  ICMP Destination unreachable (Fragmentation needed) を送信するか否かを設定する。
[ノート ]
ジャンボフレームを使用する  LAN では、ホストやスイッチングハブによってジャンボフレームの最大値が異なる。
そのため、 LAN 上に存在するすべての機器のジャンボフレームサイズをそろえておかないと通信できなくなってし
まう。
設定ミスにより、ルーターのフレームサイズより大きなパケットを送信するよう設定されたホストがあった時に、
ルーターは通常、自身のインタフェースの  MTU を越える長さのパケットを受信した場合には単にそれを破棄する
が、このコマンドを  on と設定しておくとそのようなパケットにも  ICMP エラーを返すようになる。このことにより
経路  MTU 探索が有効に働き、ホストが早めにフレームサイズを小さく切り詰めることが期待できる。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300
13.2 IPv6 の設定
13.2.1 ICMP Echo Reply を送信するか否かの設定
[書式 ]
ipv6 icmp  echo-reply  send  send
no ipv6 icmp  echo-reply  send  [send]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
[説明 ]
ICMP Echo Reply を送信するか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.2.2 ICMP Echo Reply をリンクダウン時に送信するか否かの設定
[書式 ]
ipv6 icmp  echo-reply  send-only-linkup  send
no ipv6 icmp  echo-reply  send-only-linkup  [send]
[設定値及び初期値 ]
•send
• [設定値 ] :コマンドリファレンス  | ICMP の設定  | 265

設定値 説明
on リンクアップしている時だけ  ICMP Echo Reply を返
す
off リンクの状態に関わらず  ICMP Echo Reply を返す
• [初期値 ] : off
[説明 ]
リンクダウンしているインタフェースに付与された  IP アドレスを終点  IP アドレスとする  ICMP Echo を受信した時
に、それに対して  ICMP Echo Reply を返すかどうかを設定する。 on に設定した時には、リンクアップしている時だ
け ICMP Echo を返すので、リンクの状態を  ping で調べることができるようになる。 off に設定した場合には、リン
クの状態に関わらず  ICMP Echo を返す。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.2.3 ICMP Parameter Problem を送信するか否かの設定
[書式 ]
ipv6 icmp  parameter-problem  send  send
no ipv6 icmp  parameter-problem  send  [send]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : off
[説明 ]
ICMP Parameter Problem を送信するか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.2.4 ICMP Redirect を送信するか否かの設定
[書式 ]
ipv6 icmp  redirect  send  send
no ipv6 icmp  redirect  send  [send]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
[説明 ]
ICMP Redirect を出すか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830266 | コマンドリファレンス  | ICMP の設定

13.2.5 ICMP Redirect 受信時の処理の設定
[書式 ]
ipv6 icmp  redirect  receive  action
no ipv6 icmp  redirect  receive  [action ]
[設定値及び初期値 ]
•action
• [設定値 ] :
設定値 説明
on 処理する
off 無視する
• [初期値 ] : off
[説明 ]
ICMP Redirect を受けた場合に処理するか無視するかを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.2.6 ICMP Time Exceeded を送信するか否かの設定
[書式 ]
ipv6 icmp  time-exceeded  send  send [rebound =sw]
no ipv6 icmp  time-exceeded  send  [send rebound =sw]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
•sw
• [設定値 ] :
設定値 説明
on 受信インターフェースから送信する
off 経路に従って送信する
• [初期値 ] : off
[説明 ]
ICMP Time Exceeded を出すか否かを設定する。
rebound オプションを  on に設定した場合には、経路に関係なく元となるパケットを受信したインターフェースから
送信する。
[ノート ]
rebound オプションは、 RTX5000 / RTX3500 Rev.14.00.11 以前では指定不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.2.7 ICMP Destination Unreachable を送信するか否かの設定
[書式 ]
ipv6 icmp  unreachable  send  send [rebound =sw]コマンドリファレンス  | ICMP の設定  | 267

no ipv6 icmp  unreachable  send  [send rebound =sw]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
•sw
• [設定値 ] :
設定値 説明
on 受信インターフェースから送信する
off 経路に従って送信する
• [初期値 ] : off
[説明 ]
ICMP Destination Unreachable を出すか否かを設定する。
rebound オプションを  on に設定した場合には、経路に関係なく元となるパケットを受信したインターフェースから
送信する。
[ノート ]
rebound オプションは、 RTX5000 / RTX3500 Rev.14.00.11 以前では指定不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.2.8 受信した  ICMP のログを記録するか否かの設定
[書式 ]
ipv6 icmp  log log
no ipv6 icmp  log [log]
[設定値及び初期値 ]
•log
• [設定値 ] :
設定値 説明
on 記録する
off 記録しない
• [初期値 ] : off
[説明 ]
受信した  ICMP を DEBUG タイプのログに記録するか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.2.9 ICMP Packet-Too-Big を送信するか否かの設定
[書式 ]
ipv6 icmp  packet-too-big  send  send
no ipv6 icmp  packet-too-big  send  [send]
[設定値及び初期値 ]
•send268 | コマンドリファレンス  | ICMP の設定

• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
[説明 ]
ICMP Packet-Too-Big を出すか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.2.10 IPsec で復号したパケットに対して  ICMP エラーを送るか否かの設定
[書式 ]
ipv6 icmp  error-decrypted-ipsec  send  switch
no ipv6 icmp  error-decrypted-ipsec  send  [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on IPsec で復号したパケットに対して  ICMP エラーを送
る
off IPsec で復号したパケットに対して  ICMP エラーを送
らない
• [初期値 ] : on
[説明 ]
IPsec で復号したパケットに対して  ICMP エラーを送るか否か設定する。
[ノート ]
ICMP エラーには復号したパケットの先頭部分が含まれるため、 ICMP エラーが送信元に返送される時にも  IPsec で
処理されないようになっていると、 本来  IPsec で保護したい通信が保護されずにネットワークに流れてしまう可能性
がある。特に、 フィルタ型ルーティングでプロトコルによって  IPsec で処理するかどうか切替えている場合には注意
が必要となる。
ICMP エラーを送らないように設定すると、 traceroute に対して反応がなくなるなどの現象になる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.2.11 ステルス機能の設定
[書式 ]
ipv6 stealth  all
ipv6 stealth  interface  [interface ...]
no ipv6 stealth  [...]
[設定値及び初期値 ]
• all : すべての論理インタフェースからのパケットに対してステルス動作を行う
• [初期値 ] : -
•interface
• [設定値 ] : 指定した論理インタフェースからのパケットに対してステルス動作を行う
• [初期値 ] : -
[説明 ]
このコマンドを設定すると、指定されたインタフェースから自分宛に来たパケットが原因で発生する  ICMP および
TCP リセットを返さないようになる。コマンドリファレンス  | ICMP の設定  | 269

自分がサポートしていないプロトコルや  IPv6 ヘッダ、あるいはオープンしていない  TCP/UDP ポートに対して指定
されたインタフェースからパケットを受信した時に、通常であれば  ICMP unreachable や TCP リセットを返送する。
しかし、このコマンドを設定しておくとそれを禁止することができ、ポートスキャナーなどによる攻撃を受けた時
にルーターの存在を隠すことができる。
[ノート ]
指定されたインタフェースからの  PING にも答えなくなるので注意が必要である。
自分宛ではないパケットが原因で発生する  ICMP はこのコマンドでは制御できない。それらを送信しないようにす
るには、  ipv6 icmp  * コマンド群を用いる必要がある。
ブリッジインタフェースは  RTX5000 / RTX3500 Rev.14.00.12 以降のファームウェア、および、 Rev.14.01 系以降のす
べてのファームウェアで指定可能。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
13.2.12 サイズエラーで切り詰められたフレームに対して、 ICMP エラー (Packet Too Big) を送信する
か否かの設定
[書式 ]
ipv6 icmp  packet-too-big-for-truncated  send  send
no ipv6 icmp  packet-too-big-for-truncated  send  [send]
[設定値及び初期値 ]
•send
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
[説明 ]
LAN インタフェースで受信したが、そのインタフェースの  MTU を越える長さだったために切り詰められたフレー
ムに対して  ICMP Packet Too Big を送信するか否かを設定する。
[ノート ]
ジャンボフレームを使用する  LAN では、ホストやスイッチングハブによってジャンボフレームの最大値が異なる。
そのため、 LAN 上に存在するすべての機器のジャンボフレームサイズをそろえておかないと通信できなくなってし
まう。
設定ミスにより、ルーターのフレームサイズより大きなフレームを送信するよう設定されたホストがあった時に、
ルーターは通常、自身のインタフェースの  MTU を越える長さのフレームを受信した場合には単にそれを破棄する
が、このコマンドを  on と設定しておくとそのようなフレームにも  ICMP エラーを返すようになる。このことにより
経路  MTU 探索が有効に働き、ホストが早めにフレームサイズを小さく切り詰めることが期待できる。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300270 | コマンドリファレンス  | ICMP の設定

