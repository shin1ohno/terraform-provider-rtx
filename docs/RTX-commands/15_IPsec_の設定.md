# 第15章: IPsec の設定

> 元PDFページ: 281-339

---

第 15 章
IPsec の設定
暗号化により  IP 通信に対するセキュリティを保証する  IPsec 機能を実装しています。 IPsec では、鍵交換プロトコル
IKE(Internet Key Exchange) を使用します。必要な鍵は  IKE により自動的に生成されますが、鍵の種となる事前共有
鍵は ipsec ike pre-shared-key  コマンドで事前に登録しておく必要があります。この鍵はセキュリティ・ゲートウェイ
ごとに設定できます。また、鍵交換の要求に応じるかどうかは、 ipsec ike remote address  コマンドで設定します。
鍵や鍵の寿命、暗号や認証のアルゴリズムなどを登録した管理情報は、 SA(Security Association) で管理します。 SA
を区別する  ID は自動的に付与されます。 SA の ID や状態は show ipsec sa  コマンドで確認することができます。 SA
には、鍵の寿命に合わせた寿命があります。 SA の属性のうちユーザが指定可能なパラメータをポリシーと呼びま
す。またその番号はポリシー ID と呼び、 ipsec sa policy  コマンドで定義し、 ipsec ike duration ipsec-sa 、ipsec ike
duration isakmp-sa  コマンドで寿命を設定します。
SA の削除は ipsec sa delete  コマンドで、 SA の初期化は ipsec refresh sa  コマンドで行います｡ ipsec auto refresh  コマン
ドにより､ SA を自動更新させることも可能です｡
IPsec による通信には、大きく分けてトンネルモードとトランスポートモードの  2 種類があります。
トンネルモードは  IPsec による  VPN(Virtual Private Network) を利用するためのモードです。ルーターがセキュリテ
ィ・ゲートウェイとなり、 LAN 上に流れる  IP パケットデータを暗号化して対向のセキュリティ．ゲートウェイとの
間でやりとりします。ルーターが  IPsec に必要な処理をすべて行うので、 LAN 上の始点や終点となるホストには特
別な設定を必要としません。
トンネルモードを用いる場合は、トンネルインタフェースという仮想的なインタフェースを定義し、処理すべき  IP
パケットがトンネルインタフェースに流れるように経路を設定します。個々のトンネルインタフェースはトンネル
インタフェース番号で管理されます。設定のためにトンネル番号を切替えるには tunnel select  コマンドを使用しま
す。トンネルインタフェースを使用するか使用しないかは、それぞれ tunnel enable 、tunnel disable  コマンドを使用
します。
相手先情報番号による設定トンネルインタフェース番号による
設定
•pp enable
•pp disable
•pp select<==>•tunnel enable
•tunnel disable
•tunnel select
トランスポートモードは特殊なモードであり、ルーター自身が始点または終点になる通信に対してセキュリティを
保証するモードです。ルーターからリモートのルーターへ  TELNET で入る  ( vRX シリーズ  は非対応  ) などの特殊な
場合に利用できます。トランスポートモードを使用するには ipsec transport  コマンドで定義を行い、使用をやめる
には no ipsec transport  コマンドで定義を削除します。
セキュリティ・ゲートウェイの識別子とトンネルインタフェース番号はモデルにより異なり、以下の表のようにな
ります。
モデルセキュリティ・ゲートウェイの識別
子トンネルインタフェース番号
vRX シリーズ 1-6000 (*1) 1-6000 (*1)
RTX5000 1-3000 1-3000
RTX3510 1-1000 1-1000
RTX3500 1-1000 1-1000
RTX1300 1-100 1-100
RTX1220 1-100 1-100
RTX1210 1-100 1-100
RTX840 1-20 1-20
RTX830 1-20 1-20
*1 vRX Amazon EC2 版、 vRX VMware ESXi 版 の コンパクトモード  動作時は  1-1000 となる。コマンドリファレンス  | IPsec の設定  | 281

本機はメインモード  (main mode) とアグレッシブモード  (aggressive mode) に対応しています。 VPN を構成する両方
のルーターが固定のグローバルアドレスを持つときにはメインモードを使用し、一方のルーターしか固定のグロー
バルアドレスを持たないときにはアグレッシブモードを使用します。
メインモードを使用するためには、 ipsec ike remote address  コマンドで対向のルーターの  IP アドレスを設定する必
要があります。アグレッシブモードを使用するときには、固定のグローバルアドレスを持つかどうかによって設定
が異なります。固定のグローバルアドレスを持つルーターには、 ipsec ike remote name  コマンドを設定し、 ipsec ike
remote address  コマンドで  any を設定します。固定のグローバルアドレスを持たないルーターでは、 ipsec ike local
name  コマンドを設定し、 ipsec ike remote address  コマンドで  IP アドレスを設定します。
メインモードでは、 ipsec ike local name  コマンドや ipsec ike remote name  コマンドを設定することはできません。ま
た、アグレッシブモードでは、 ipsec ike local name  コマンドと ipsec ike remote name  コマンドの両方を同時に設定す
ることはできません。このように設定した場合には、正しく動作しない可能性があります。
15.1 IPsec の動作の設定
[書式 ]
ipsec  use use
no ipsec  use [use]
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on 動作させる
off 動作させない
• [初期値 ] : on
[説明 ]
IPsec を動作させるか否かを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.2 IKE バージョンの設定
[書式 ]
ipsec  ike version  gateway_id  version
no ipsec  ike version  gateway_id  [version ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•version
• [設定値 ] : 使用する IKEのバージョン
• [設定値 ] :
設定値 説明
1 IKEバージョン 1
2 IKEバージョン 2
• [初期値 ] : 1
[説明 ]
セキュリティ・ゲートウェイで使用する IKEのバージョンを設定する。282 | コマンドリファレンス  | IPsec の設定

[ノート ]
versionで指定したバージョン以外での接続以外は受け付けない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.3 IKE の認証方式の設定
[書式 ]
ipsec  ike auth  method  gateway_id  method
no ipsec  ike auth  method  gateway_id  [method ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•method
• [設定値 ] :
設定値 説明
auto 認証方式を自動的に選択する
pre-shared-key 事前共有鍵
certificate デジタル署名
eap-md5 EAP-MD5
• [初期値 ] :
• auto
[説明 ]
IKEの認証方式を設定する。
METHOD にautoを設定した場合、以下の条件にしたがって認証方式が決定される。
•事前共有鍵方式
•ipsec ike pre-shared-key  コマンドが設定されている場合。
•デジタル署名方式
次の条件をすべて満たしている場合
•ipsec ike pki file  コマンドで指定した場所に証明書が保存されている。
•ipsec ike eap request  コマンドおよび  ipsec ike eap myname  コマンドが設定されていない。
• EAP-MD5 方式
次の条件をすべて満たしている場合
•ipsec ike pki file  コマンドで指定した場所に証明書が保存されている。
•ipsec ike eap request  コマンド、または  ipsec ike eap myname  コマンドが設定されていない。
上記、認証方式を決定する条件のうち、複数の条件に合致する場合、次の順番で認証方式が優先される。コマンドリファレンス  | IPsec の設定  | 283

1.事前共有鍵方式
2.デジタル署名方式
3.EAP-MD5 方式
method  に auto 以外を指定した場合、上記の認証方式を決定する条件にかかわらず、  method  に指定した方式で認証
を行う。
[ノート ]
本コマンドは  IKEv2 でのみ有効であり、  IKEv1 の動作に影響を与えない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.4 IKE で受信した  ID を基とした経路を自動的に追加するか否かの設定
[書式 ]
ipsec  ike auto  received-id-route-add  gateway_id  switch
no ipsec  ike auto  received-id-route-add  gateway_id  [switch ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on 経路を自動的に追加する
off 経路を自動的に追加しない
• [初期値 ] : off
[説明 ]
セキュリティ・ゲートウェイの  Responder として動作する場合に、 IKEv1 のフェーズ  2 でセキュリティ・ゲートウェ
イの  Initiator から受信した  ID(IPv4 アドレス、または、 IPv4 ネットワークアドレス ) を基にした経路を、自動的に追
加するか否かを設定する。
本コマンドは、トンネルが  IPsec のトンネルモードで動作しており、かつ、鍵交換に  IKEv1 を使用し、 IKE のネゴシ
エーションをアグレッシブモードで行っているときのみ有効となる。
経路の追加  / 削除は  IPsec トンネルの  UP / DOWN に連動しており、送信用の  IPsec SA が確立したときに経路が追加
され、送信用の IPsec SA が削除されると経路も削除される。
[ノート ]
IKEv1 のフェーズ  2 で ID の交換を行わなかった場合は経路を追加しない。 IKEv1 のフェーズ  2 で交換した  ID が
IPv6 であった場合は経路を追加しない。本コマンドの設定を変更すると、 SAが削除され、 IKE の状態が初期化され
る。本コマンドは  IKEv1 でのみ有効であり、 IKEv2 の動作に影響を与えない。本コマンドはトンネルモードかつア
グレッシブモードで有効であり、トランスポートモードの動作およびメインモードの動作に影響を与えない。284 | コマンドリファレンス  | IPsec の設定

セキュリティ・ゲートウェイの  Initiator として使用するルーターには、 IKEv1 のフェーズ  2 で ID を送信するために
ローカル  ID とリモート  ID の設定を以下のいずれかの方法で行う必要がある。なお、マルチポイントトンネルでは
1 の方法で設定する必要がある。
1. ipsec ike local id  コマンド、および、 ipsec ike remote id  コマンドの両コマンドで設定する
2. ipsec sa policy  コマンドの  local-id  パラメーター、および、 remote-id  パラメーターの両パラメーターで設定する
ipsec sa policy  コマンドの  local-id  / remote-id  パラメーターと  ipsec ike local id  / ipsec ike remote id  コマンドの両方に
ID が設定されている場合は、 ipsec sa policy  コマンドの  local-id  / remote-id  パラメーターに設定した  ID が IKEv1 のフ
ェーズ 2で送信される。
RTX1300 Rev.23.00.08 以前では使用不可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX さくらのクラウド版 , RTX3510, RTX1300
15.5 事前共有鍵の登録
[書式 ]
ipsec  ike pre-shared-key  gateway_id  key
ipsec  ike pre-shared-key  gateway_id  text text
no ipsec  ike pre-shared-key  gateway_id  [...]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•key
• [設定値 ] : 鍵となる  0x ではじまる十六進数列  (128バイト以内  )
• [初期値 ] : -
•text
• [設定値 ] : ASCII 文字列で表した鍵  (128文字以内  )
• [初期値 ] : -
[説明 ]
鍵交換に必要な事前共有鍵を登録する。設定されていない場合には、鍵交換は行われない。
鍵交換を行う相手ルーターには同じ事前共有鍵が設定されている必要がある。
[設定例 ]
ipsec ike pre-shared-key 1 text himitsu
ipsec ike pre-shared-key 8 0xCDEEEDC0CDEDCD
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_idコマンドリファレンス  | IPsec の設定  | 285

ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.6 IKEv2 の認証に使用する  PKI ファイルの設定
[書式 ]
ipsec  ike pki file gateway_id  certificate =cert_id  [crl=crl_id ]
no ipsec  ike pki file gateway_id  [...]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•cert_id
• [設定値 ] :
設定値 説明
1..8 証明書ファイルの識別子
• [初期値 ] : -
•crl_id
• [設定値 ] :
設定値 説明
1..8 CRLファイルの識別子
• [初期値 ] : -
[説明 ]
IKEv2 の認証に使用する PKIファイルを設定する。
デジタル証明書方式の認証を行う場合、  cert_id  に使用する証明書が保存されているファイルの識別子を指定する。
EAP-MD5 認証を行う場合、始動側は相手の証明書を検証するために  cert_id  に自分の証明書が保存されているファ
イルの識別子を指定する。
[ノート ]
本コマンドは  IKEv2 でのみ有効であり、  IKEv1 の動作に影響を与えない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -286 | コマンドリファレンス  | IPsec の設定

[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.7 EAP-MD5 認証で使用する自分の名前とパスワードの設定
[書式 ]
ipsec  ike eap myname  gateway_id  name  password
no ipsec  ike eap myname  gateway_id  [...]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•name
• [設定値 ] : 名前  (半角 256文字以内 )
• [初期値 ] : -
•password
• [設定値 ] : パスワード (半角 64文字以内 )
• [初期値 ] : -
[説明 ]
EAP-MD5 認証を要求されたときに使用する名前とパスワードを設定する。
[ノート ]
本コマンドは  IKEv2 でのみ有効であり、 IKEv1 の動作に影響を与えない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.8 EAP-MD5 によるユーザ認証の設定
[書式 ]
ipsec  ike eap request  gateway_id  sw group_id
no ipsec  ike eap request  gateway_id  [...]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•sw
• [設定値 ] :
設定値 説明
on 要求するコマンドリファレンス  | IPsec の設定  | 287

設定値 説明
off 要求しない
• [初期値 ] : off
•group_id
• [設定値 ] : 認証に使用するユーザグループの識別番号
• [初期値 ] : -
[説明 ]
IKEv2 で、EAP-MD5 認証をクライアントに要求するか否かを設定する。  group_id  を指定した場合には、該当のユー
ザグループに含まれるユーザを認証の対象とする。
本コマンドによる設定はルーターが応答側として動作するときにのみ有効であり、始動側のセキュリティゲートウ
ェイから送信された  IKE AUTH 交換に  AUTH ペイロードが含まれない場合に  EAP-MD5 によるユーザ認証を行う。
[ノート ]
本コマンドは  IKEv2 でのみ有効であり、 IKEv1 の動作に影響を与えない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
•groupid
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX2 1000 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.9 EAP-MD5 認証で証明書要求ペイロードを送信するか否かの設定
[書式 ]
ipsec  ike eap send  certreq  gateway_id  switch
no ipsec  ike eap send  certreq  gateway_id  [switch ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない288 | コマンドリファレンス  | IPsec の設定

• [初期値 ] : off
[説明 ]
EAP-MD5 認証方式の場合、始動側のセキュリティ・ゲートウェイから送信する IKE_AUTH 交換に、証明書要求
（CERTREQ ）ペイロードを含めるか否かを設定する。
[ノート ]
本コマンドは  IKEv2 でのみ有効であり、 IKEv1 の動作に影響を与えない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.10 IKE の鍵交換を始動するか否かの設定
[書式 ]
ipsec  auto  refresh  [gateway_id ] switch
no ipsec  auto  refresh  [gateway_id ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on 鍵交換を始動する
off 鍵交換を始動しない
• [初期値 ] :
• off ( 全体的な動作  )
• on ( gateway_id  毎 )
[説明 ]
IKE の鍵交換を始動するかどうかを設定する。他のルーターが始動する鍵交換については、このコマンドに関係な
く常に受け付ける。
gateway_id  パラメータを指定しない書式は、ルーターの全体的な動作を決める。この設定が  off のときにはルーター
は鍵交換を始動しない。
gateway_id  パラメータを指定する書式は、指定したセキュリティゲートウェイに対する鍵交換の始動を抑制するた
めに用意されている。
例えば、次の設定では、 1 番のセキュリティゲートウェイのみが鍵交換を始動しない。
ipsec auto refresh on
ipsec auto refresh 1 offコマンドリファレンス  | IPsec の設定  | 289

[ノート ]
ipsec auto refresh  off の設定では、 gateway_id  パラメータを指定する書式は効力を持たない。例えば、次の設定では、
1 番のセキュリティゲートウェイでは鍵交換を始動しない。
ipsec auto refresh off ( デフォルトの設定  )
ipsec auto refresh 1 on
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.11 設定が異なる場合に鍵交換を拒否するか否かの設定
[書式 ]
ipsec  ike negotiate-strictly  gateway_id  switch
no ipsec  ike negotiate-strictly  gateway_id
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on 鍵交換を拒否する
off 鍵交換を受理する
• [初期値 ] : off
[説明 ]
IKEv1 として動作する際、 設定が異なる場合に鍵交換を拒否するか否かを設定する。  このコマンドの設定が  off のと
きには、従来のリビジョンと同様に動作する。すなわち、相手の提案するパラメータが自分の設定と異なる場合で
も、そのパラメータをサポートしていれば、それを受理する。このコマンドの設定が  on のときには、同様の状況で
相手の提案を拒否する。このコマンドが適用されるパラメータと対応するコマンドは以下の通りである。
パラメータ 対応するコマンド
暗号アルゴリズム ipsec ike encryption
グループ ipsec ike group
ハッシュアルゴリズム ipsec ike hash
PFS ipsec ike pfs
フェーズ  1 のモード ipsec ike local name  など
[ノート ]
本コマンドは  IKEv2 としての動作には影響を与えない。290 | コマンドリファレンス  | IPsec の設定

[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.12 IKE の鍵交換に失敗したときに鍵交換を休止せずに継続するか否かの設定
[書式 ]
ipsec  ike always-on  gateway_id  switch
no ipsec  ike always-on
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on 鍵交換を継続する
off 鍵交換を休止する
• [初期値 ] : off
[説明 ]
IKE の鍵交換に失敗したときに鍵交換を休止せずに継続できるようにする。 IKE キープアライブを用いるときには、
このコマンドを設定しなくても、常に鍵交換を継続する。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.13 鍵交換の再送回数と間隔の設定
[書式 ]
ipsec  ike retry  count  interval  [max_session ]
no ipsec  ike retry  [count  interval  [max_session ]]コマンドリファレンス  | IPsec の設定  | 291

[設定値及び初期値 ]
•count
• [設定値 ] : 再送回数  (1..50)
• [初期値 ] : 10
•interval
• [設定値 ] : 再送間隔の秒数  (1..100)
• [初期値 ] : 5
•max_session
• [設定値 ] : 同時に動作するフェーズ  1 の最大数  (1..5)
• [初期値 ] : 3
[説明 ]
鍵交換のパケットが相手に届かないときに実施する再送の回数と間隔を設定する。
また、 max_session パラメータは、  IKEv1 において同時に動作するフェーズ  1 の最大数を指定する。ルーターは、フ
ェーズ  1 が確立せずに再送を継続する状態にあるとき、鍵の生成を急ぐ目的で、新しいフェーズ  1 を始動すること
がある。このパラメータは、このような状況で、同時に動作するフェーズ  1 の数を制限するものである。なお、こ
のパラメータは、始動側のフェーズ  1 のみを制限するものであり、応答側のフェーズ  1 に対しては効力を持たな
い。
[ノート ]
IKEv2 として動作する場合、 max_session  パラメータは効力を持たない。同じ相手側セキュリティ・ゲートウェイに
対して始動する鍵交換セッションは、常に最大  1 セッションとなる。
相手側セキュリティ・ゲートウェイに掛かっている負荷が非常に高い場合、本コマンドの設定値を調整することで
鍵交換が成功しやすくなる可能性がある。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.14 相手側のセキュリティ・ゲートウェイの名前の設定
[書式 ]
ipsec  ike remote  name  gateway  name  [type]
no ipsec  ike remote  name  gateway  [name ]
[設定値及び初期値 ]
•gateway
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•name
• [設定値 ] : 名前  ( 256文字以内  )
• [初期値 ] : -
•type : id の種類
• [設定値 ] :
設定値 説明
ipv4-addr ID_IPV4_ADDR
fqdn ID_FQDN
user-fqdn( もしくは rfc822-addr) ID_USER_FQDN(ID_RFC822_ADDR)
ipv6-addr ID_IPV6_ADDR
key-id ID_KEY_ID
tel NGN 網電話番号  (ID_IPV6_ADDR)
( vRX Amazon EC2 版、および  vRX さくらのクラウド
版 では指定不可能  )292 | コマンドリファレンス  | IPsec の設定

設定値 説明
tel-key NGN 網電話番号  (ID_KEY_ID)
( vRX Amazon EC2 版、および  vRX さくらのクラウド
版 では指定不可能  )
• [初期値 ] : -
[説明 ]
相手側のセキュリティ・ゲートウェイの名前と IDの種類を設定する。
その他、動作する IKEのバージョンによって異なる、本コマンドの影響、注意点については以下の通り。
• IKEv1
このコマンドの設定は、フェーズ  1 のアグレッシブモードで使用され、メインモードでは使用されない。
また、  type パラメータは相手側セキュリティ・ゲートウェイの判別時に考慮されない。
• IKEv2
相手側セキュリティ・ゲートウェイの判別時には  name  、 type パラメータの設定が共に一致している必要がある。
type パラメータが  'tel' の場合、相手側 IPv6アドレス (ID_IPV6_ADDR) を相手側セキュリティ・ゲートウェイの判
別に使用する。
type パラメータが  'tel-key' の場合、設定値を ID_KEY_ID として相手側セキュリティ・ゲートウェイの判別に使用
する。
type パラメータが  'key-id' 以外の場合、  name  から相手側セキュリティ・ゲートウェイの  IP アドレスの特定を試
み、特定できれば、そのホストに対して鍵交換を始動する。この場合、 ipsec ike remote address  コマンドの設定
は不要である。
ただし、 ipsec ike remote address  コマンドが設定されている場合は、そちらの設定にしたがって始動時の接続先
ホストが決定される。
[ノート ]
'tel'および 'tel-key'は、データコネクト拠点間接続機能で使用する。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.15 相手側セキュリティ・ゲートウェイの  IP アドレスの設定
[書式 ]
ipsec  ike remote  address  gateway_id  ip_address
no ipsec  ike remote  address  gateway_id  [ip_address ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•ip_address
• [設定値 ] :コマンドリファレンス  | IPsec の設定  | 293

設定値 説明
IP アドレス、またはホスト名相手側セキュリティ・ゲートウェイの  IP アドレス、
またはホスト名 (半角 255文字以内 )
any 自動選択
• [初期値 ] : -
[説明 ]
相手側セキュリティ・ゲートウェイの  IP アドレスまたはホスト名を設定する。ホスト名で設定した場合には、鍵交
換の始動時にホスト名から  IP アドレスを  DNS により検索する。
その他、動作する IKEバージョンによって異なる、本コマンドの影響、注意点については以下の通り。
• IKEv1
応答側になる場合、本コマンドで指定したホストは相手側セキュリティ･ゲートウェイの判別に使用される。
'any' が設定された場合は、相手側セキュリティ・ゲートウェイとして任意のホストから鍵交換を受け付ける。そ
の代わりに、自分から鍵交換を始動することはできない。例えば、アグレッシブモードで固定のグローバルアド
レスを持つ場合などに利用する。
• IKEv2
このコマンドで設定したホストは、 鍵交換を始動する際の接続先としてのみ使用される。  'any' は自分側から鍵交
換を始動しないことを明示的に示す。
応答側となる場合、本コマンドの設定による相手側セキュリティ・ゲートウェイの判別は  ipsec ike remote name
コマンド等の設定によって行われる。
[ノート ]
ホスト名を指定する場合には、  dns server  コマンドなどで必ず  DNS サーバーを設定しておくこと。  IPsecメインモ
ード接続では、相手側セキュリティ・ゲートウェイの  IP アドレスおよびホスト名を重複して設定しない。  相手側セ
キュリティ・ゲートウェイの  IP アドレスおよびホスト名を重複して設定した場合の動作は保証されない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.16 相手側の  ID の設定
[書式 ]
ipsec  ike remote  id gateway_id  ip_address [/mask ]
no ipsec  ike remote  id gateway_id  [ip_address [/mask ]]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•ip_address
• [設定値 ] : IP アドレス
• [初期値 ] : -
•mask294 | コマンドリファレンス  | IPsec の設定

• [設定値 ] : ネットマスク
• [初期値 ] : -
[説明 ]
IKEv1 のフェーズ  2 で用いる相手側の  ID を設定する。
このコマンドが設定されていない場合は、フェーズ  2 で ID を送信しない。
mask  パラメータを省略した場合は、タイプ  1 の ID が送信される。また、 mask  パラメータを指定した場合は、タイ
プ 4 の ID が送信される。
[ノート ]
本コマンドは  IKEv2 の動作には影響を与えない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.17 自分側のセキュリティ・ゲートウェイの名前の設定
[書式 ]
ipsec  ike local  name  gateway_id  name  [type]
no ipsec  ike local  name  gateway_id  [name ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•name
• [設定値 ] : 名前  ( 256文字以内  )
• [初期値 ] : -
•type : id の種類
• [設定値 ] :
設定値 説明
ipv4-addr ID_IPV4_ADDR
fqdn ID_FQDN
user-fqdn( もしくは rfc822-addr) ID_USER_FQDN (ID_RFC822_ADDR)
ipv6-addr ID_IPV6_ADDR
key-id ID_KEY_ID
tel NGN 網電話番号  (ID_IPV6_ADDR)
( vRX Amazon EC2 版、および  vRX さくらのクラウド
版 では指定不可能  )コマンドリファレンス  | IPsec の設定  | 295

設定値 説明
tel-key NGN 網電話番号  (ID_KEY_ID)
( vRX Amazon EC2 版、および  vRX さくらのクラウド
版 では指定不可能  )
• [初期値 ] : -
[説明 ]
自分側のセキュリティ・ゲートウェイの名前と  ID の種類を設定する。
なお、 IKEv1として動作する際に  type パラメータが  'ipv4-addr' 、 'ipv6-addr' 、'tel'、'tel-key' に設定されていた場合は
'key-id' を設定したときと同等の動作となる。  IKEv2かつ  type パラメータが  'tel' の場合、自分側 IPv6アドレス
(ID_IPV6_ADDR) を鍵交換に使用する。  IKEv2かつ  type パラメータが  'tel-key' の場合、設定値を ID_KEY_ID として
鍵交換に使用する。
[ノート ]
'tel'および 'tel-key'は、データコネクト拠点間接続機能で使用する。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.18 自分側セキュリティ・ゲートウェイの  IP アドレスの設定
[書式 ]
ipsec  ike local  address  gateway_id  ip_address
ipsec  ike local  address  gateway_id  vrrp interface  vrid
ipsec  ike local  address  gateway_id  ipv6 prefix prefix  on interface
ipsec  ike local  address  gateway_id  ipcp pp peer_num
no ipsec  ike local  address  gateway_id  [ip_address ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•ip_address
• [設定値 ] : 自分側セキュリティ･ゲートウェイの  IP アドレス
• [初期値 ] : -
•interface
• [設定値 ] : LAN インタフェース名
• [初期値 ] : -
•vrid
• [設定値 ] : VRRP グループ  ID(1..255)
• [初期値 ] : -
•prefix
• [設定値 ] : プレフィックス296 | コマンドリファレンス  | IPsec の設定

• [初期値 ] : -
•peer_num
• [設定値 ] : PP インタフェース番号
• [初期値 ] : -
[説明 ]
自分側セキュリティ・ゲートウェイの  IP アドレスを設定する。
vrrp キーワードを指定する第  2 書式では、 VRRP マスターとして動作している場合のみ、指定した  LAN インタフェ
ース /VRRP グループ  ID の仮想  IP アドレスを自分側セキュリティ・ゲートウェイアドレスとして利用する。
VRRP マスターでない場合には鍵交換は行わない。
ipv6 キーワードを指定する第  3 書式では、 IPv6 のダイナミックアドレスを指定する。
ipcp キーワードを指定する第  4 書式では、 IPCP アドレスを取得する  PP インタフェースを指定する。
[ノート ]
本コマンドが設定されていない場合には、相手側のセキュリティ・ゲートウェイに近いインタフェースの  IP アドレ
スを用いて  IKE を起動する。
第 2 書式は、 vRX Amazon EC2 版、および  vRX さくらのクラウド版  では使用不可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
•peer_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 150
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.19 自分側の  ID の設定
[書式 ]
ipsec  ike local  id gateway_id  ip_address [/mask ]
no ipsec  ike local  id gateway_id  [ip_address [/mask ]]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•ip_address
• [設定値 ] : IP アドレス
• [初期値 ] : -
•maskコマンドリファレンス  | IPsec の設定  | 297

• [設定値 ] : ネットマスク
• [初期値 ] : -
[説明 ]
IKEv1 のフェーズ  2 で用いる自分側の  ID を設定する。
このコマンドが設定されていない場合には、フェーズ  2 で ID を送信しない。
mask  パラメータを省略した場合は、タイプ  1 の ID が送信される。
また、 mask  パラメータを指定した場合は、タイプ  4 の ID が送信される。
[ノート ]
本コマンドは  IKEv2 としての動作には影響を与えない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.20 IKE キープアライブ機能の設定
[書式 ]
ipsec  ike keepalive  use gateway_id  switch  [down=disconnect] [send-only-new-sa= send]
ipsec  ike keepalive  use gateway_id  switch  heartbeat [ interval  count  [upwait ]] [down=disconnect] [send-only-new-sa= send]
ipsec  ike keepalive  use gateway_id  switch  icmp-echo ip_address  [length= length ] [interval  count  [upwait ]]
[down=disconnect]
ipsec  ike keepalive  use gateway_id  switch  dpd [ interval  count  [upwait ]]
ipsec  ike keepalive  use gateway_id  switch  rfc4306 [ interval  count  [upwait ]]
no ipsec  ike keepalive  use gateway_id  [switch  ....]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•switch  : キープアライブの動作
• [設定値 ] :
設定値 説明
on キープアライブを使用する
off キープアライブを使用しない
auto 対向のルーターがキープアライブを送信するときに
限って送信する  (heartbeat 、 rfc4306 でのみ有効  )
• [初期値 ] : auto
•ip_address
• [設定値 ] : ping を送信する宛先の  IP アドレス  (IPv4/IPv6)
• [初期値 ] : -
•length298 | コマンドリファレンス  | IPsec の設定

• [設定値 ] : ICMP echo のデータ部の長さ  (64..1500)
• [初期値 ] : 64
•interval
• [設定値 ] : キープアライブパケットの送信間隔秒数  (1..600)
• [初期値 ] : 10
•count
• [設定値 ] : キープアライブパケットが届かないときに障害とみなすまでの試行回数  (1..50)
• [初期値 ] : 6
•upwait
• [設定値 ] : IPsec SA が生成されてから実際にトンネルインタフェースを有効にするまでの時間  (0..1000000)
• [初期値 ] : 0
•send
• [設定値 ] :
設定値 説明
on 新旧の  SA が混在する場合、新しい  SA のみに対して
キープアライブパケットを送信する
off 新旧の  SA が混在する場合、新旧  SA の両方に対して
キープアライブパケットを送信する
auto 新旧の  SA が混在する場合、対向からの古い  SA に対
するキープアライブパケットを受信した場合のみ新
旧 SA の両方に対してキープアライブパケットを送
信し、そうでない場合は新しい  SA のみに対してキー
プアライブパケットを送信する
• [初期値 ] :
• auto ( vRX さくらのクラウド版、 RTX1300 、RTX3510 、RTX840 )
• off ( 上記以外  )
[説明 ]
IKE キープアライブの動作を設定する。
本コマンドは、動作する IKEのバージョンによって以下のように動作が異なる。
• IKEv1
キープアライブの方式としては、 heartbeat、ICMP Echo 、DPD(RFC3706) の 3 種類から選ぶことができる。第  1 書
式は自動的に  heartbeat 書式となる。
heartbeat 書式を利用するには第  1 、 第 2 書式を使用する。  heartbeat 方式において  switch  パラメータが  auto に設定
されている場合は、相手から  heartbeat パケットを受信したときだけ  heartbeat パケットを送信する。従って、双方
の設定が  auto になっているときには、  IKE キープアライブは動作しない。
ICMP Echo を利用するときには第  3 書式を使用し、送信先の  IP アドレスを設定する。オプションとして、  ICMP
Echo のデータ部の長さを指定することができる。この方式では、 switch  パラメータが  auto でも  on の場合と同様
に動作する。
DPD を利用するときには第  4 書式を使用する。この方式では  switch  パラメータが  auto でも  on の場合と同様に
動作する。
その他、  IKEv1 で対応していない方式  ( 書式  ) が設定されている場合は、代替方式として  heartbeat で動作する。
このとき、  switch  、 count  、 interval  、 upwait  パラメータは設定内容が反映される。
• IKEv2
キープアライブの方式として、 RFC4306(IKEv2 標準  ) 、 ICMP Echo の 2 種類から選ぶことができる。第  1 書式は
自動的に  RFC4306 方式となる。
switch  パラメータが  auto の場合には、  RFC4306 方式のキープアライブパケットを受信したときだけ応答パケット
を送信する。なお、  IKEv2 ではこの方式のキープアライブパケットには必ず応答しなければならないため、
switch  パラメータが  auto でも  off の場合でも同様に動作する。
ICMP Echo を利用するときには第  3 書式を使用し、送信先の IPアドレスを設定する。  オプションとして、 ICMP
Echo のデータ部の長さを指定することができる。この方式では、 switch  パラメータが  auto でも  on の場合と同様
に動作する。コマンドリファレンス  | IPsec の設定  | 299

その他、  IKEv2 で対応していない方式  ( 書式  ) が設定されている場合は、代替方式として  RFC4306 で動作する。
このとき、  switch  、 count  、 interval  、 upwait  パラメータは設定内容が反映される。
[ノート ]
相手先が  PP インタフェースの先にある場合、 down オプションを指定することができる。 down オプションを指定す
ると、 キープアライブダウン検出時と  IKE の再送回数満了時に  PP インタフェースの切断を行うことができる。網側
の状態などで  PP インタフェースの再接続によりトンネル確立状態の改善を望める場合に利用することができる。
キープアライブの方式として  heartbeat を使用する場合、 send-only-new-sa オプションを指定することができる。 send-
only-new-sa オプションに  on を設定すると、鍵交換後の新旧の SAが混在するときに新しい  SA のみに対してキープ
アライブパケットを送信するようになり、鍵交換時の負荷を軽減することができる。 auto を設定すると、対向から
の古い  SA に対するキープアライブパケットを受信した場合のみこちらからも古い  SA に対してキープアライブパ
ケットを送信し、そうでない場合は  on を設定したときと同様に動作する。
send-only-new-sa オプションに対応していない  ヤマハルーター  とトンネルを構築する場合は、 send-only-new-sa オプ
ションを  on に設定しているとトンネルがダウンするため、 auto または  off に設定しておかなければならない。
length  パラメータで指定するのは  ICMP データ部分の長さであり、 IP パケット全体の長さではない。
同じ相手に対して、複数の方法を併用することはできない。
send-only-new-sa オプションは、 RTX5000 / RTX3500 Rev.14.00.28 以前では使用不可能。また  auto の指定は、 vRX
Amazon EC2 版、 vRX VMware ESXi 版、 RTX5000 、RTX3500 、RTX1220 、RTX1210 、および  RTX830 では不可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.21 IKE キープアライブに関する  SYSLOG を出力するか否かの設定
[書式 ]
ipsec  ike keepalive  log gateway_id  log
no ipsec  ike keepalive  log gateway_id  [log]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•log
• [設定値 ] :
設定値 説明
on 出力する
off 出力しない
• [初期値 ] : on
[説明 ]
IKE キープアライブに関する  SYSLOG を出力するか否かを設定する。この  SYSLOG は DEBUG レベルの出力であ
る。300 | コマンドリファレンス  | IPsec の設定

[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.22 IKE が用いる暗号アルゴリズムの設定
[書式 ]
ipsec  ike encryption  gateway_id  algorithm
no ipsec  ike encryption  gateway_id  [algorithm ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•algorithm
• [設定値 ] :
設定値 説明
3des-cbc 3DES-CBC
des-cbc DES-CBC
aes-cbc AES-CBC
aes256-cbc AES256-CBC
• [初期値 ] : 3des-cbc
[説明 ]
IKE が用いる暗号アルゴリズムを設定する。
始動側として働く場合に、本コマンドで設定されたアルゴリズムを提案する。応答側として働く場合は本コマンド
の設定に関係なく、サポートされている任意のアルゴリズムを用いることができる。
ただし、 IKEv1 で ipsec ike negotiate-strictly  コマンドが  on の場合は、応答側であっても設定したアルゴリズムしか
利用できない。
[ノート ]
IKEv2 では、 ipsec ike proposal-limitation  コマンドが  on に設定されているとき、本コマンドで設定されたアルゴリズ
ムを提案する。 ipsec ike proposal-limitation  コマンドが  off に設定されているとき、または、 ipsec ike proposal-
limitation  コマンドに対応していない機種では、本コマンドの設定にかかわらず、サポートするすべてのアルゴリズ
ムを同時に提案し、相手側セキュリティ・ゲートウェイに選択させる。また応答側として働く場合は、提案された
ものからより安全なアルゴリズムを選択する。
IKEv2 でサポート可能な暗号アルゴリズム及び応答時の選択の優先順位は以下の通り。
• AES256-CBC ＞ AES192-CBC ＞ AES128-CBC ＞ 3DES-CBC ＞ DES-CBC
※IKEv2 でのみ  AES192-CBC をサポートする。ただし、コマンドで  AES192-CBC を選択することはできない。
[設定例 ]
# ipsec ike encryption 1 aes-cbcコマンドリファレンス  | IPsec の設定  | 301

[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.23 受信した  IKE パケットを蓄積するキューの長さの設定
[書式 ]
ipsec  ike queue  length  length
no ipsec  ike queue  length  [length ]
[設定値及び初期値 ]
•length  : キュー長
• [設定値 ] :
設定値 機種
6000...24000vRX Amazon EC2 版 ( 通常モード  での動作時  ) 、vRX
VMware ESXi 版 ( 通常モード  での動作時  ) 、vRX さ
くらのクラウド版
3000...12000 RTX3510 、RTX5000
1100...4400 RTX1300 Rev.23.00.09 以降
1000...4000vRX Amazon EC2 版 ( コンパクトモード  での動作
時 ) 、vRX VMware ESXi 版 ( コンパクトモード  での
動作時  ) 、RTX3500
100....200RTX1300 Rev.23.00.05 以前、 RTX1220 、RTX1210 、
RTX830 Rev.15.02.22 以降
20..40 RTX840、RTX830 Rev.15.02.20 以前
• [初期値 ] :
• 12000 ( vRX Amazon EC2 版 ( 通常モード  での動作時  ) 、vRX VMware ESXi 版 ( 通常モード  での動作時  ) 、
vRX さくらのクラウド版  )
• 6000 ( RTX3510 、RTX5000 )
• 2200 ( RTX1300 Rev.23.00.09 以降  )
• 2000 ( vRX Amazon EC2 版 ( コンパクトモード  での動作時  ) 、vRX VMware ESXi 版 ( コンパクトモード  で
の動作時  ) 、RTX3500 )
• 200 ( RTX1300 Rev.23.00.05 以前、 RTX1220 、RTX1210 、RTX830 Rev.15.02.22 以降  )
• 40 ( RTX840 、RTX830 Rev.15.02.20 以前  )
[説明 ]
受信した  IKE パケットを蓄積するキューの長さを設定する。  この設定は、短時間に集中して  IKE パケットを受信し
た際のルーターの振る舞いを決定する。設定した値が大きいほど、 IKE パケットが集中したときにより多くのパケ
ットを取りこぼさないで処理することができるが、逆に  IKE パケットがルーターに滞留する時間が長くなるためキ
ープアライブの応答が遅れ、トンネルの障害を間違って検出する可能性が増える。  通常の運用では、この設定を変
更する必要はないが、多数のトンネルを構成しており、多数の  SA を同時に消す  状況があるならば値を大きめに設
定するとよい。302 | コマンドリファレンス  | IPsec の設定

[ノート ]
キューの長さを長くすると、一度に受信して処理できる  IKE パケットの数を増やすことができる。しかし、あまり
大きくすると、ルーター内部にたまった  IKE パケットの処理が遅れ、対向のルーターでタイムアウトと検知されて
しまう可能性が増える。そのため、このコマンドの設定を変更する時には、慎重に行う必要がある。
通常の運用では、この設定を変更する必要はない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.24 IKE が用いるグループの設定
[書式 ]
ipsec  ike group  gateway_id  group  [group ]
no ipsec  ike group  gateway_id  [group  [group ]]
[設定値及び初期値 ]
•gateway_id  : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•group  : グループ識別子
• [設定値 ] :
• modp768
• modp1024
• modp1536
• modp2048
• [初期値 ] :
• modp1024
[説明 ]
IKE で用いるグループを設定する。
始動側として働く場合には、このコマンドで設定されたグループを提案する。応答側として働く場合には、このコ
マンドの設定に関係なく、サポート可能な任意のグループを用いることができる。
その他、動作する  IKE のバージョンによって異なる本コマンドの影響、注意点については以下の通り。
• IKEv1
2 種類のグループを設定した場合には、 1 つ目がフェーズ  1 で、 2 つ目がフェーズ  2 で提案される。グループを  1
種類しか設定しない場合は、フェーズ  1 とフェーズ  2 の両方で、設定したグループが提案される。
また、  ipsec ike negotiate-strictly  コマンドが  on の場合は、応答側であっても設定したグループしか利用できな
い。
• IKEv2
常に  1 つ目に設定したグループのみが使用される。  2 つ目に設定したグループは無視される。
また、  始動側として提案したグループが相手に拒否され、別のグループを要求された場合は、そのグループで再
度提案する  ( 要求されたグループがサポート可能な場合  ) 。 以後、 IPsec の設定を変更するか再起動するまで、同
じ相手側セキュリティ・ゲートウェイに対しては再提案したグループが優先的に使用される。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -コマンドリファレンス  | IPsec の設定  | 303

[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.25 IKE が用いるハッシュアルゴリズムの設定
[書式 ]
ipsec  ike hash  gateway_id  algorithm
no ipsec  ike hash  gateway_id  [algorithm ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•algorithm
• [設定値 ] :
設定値 説明
md5 MD5
sha SHA-1
sha256 SHA-256
sha384 SHA-384
sha512 SHA-512
• [初期値 ] : sha
[説明 ]
IKE が用いるハッシュアルゴリズムを設定する。
始動側として働く場合に、本コマンドで設定されたアルゴリズムを提案する。応答側として働く場合は本コマンド
の設定に関係なく、サポートされている任意のアルゴリズムを用いることができる。
ただし、 IKEv1 で　 ipsec ike negotiate-strictly  コマンドが  on の場合は、応答側であっても設定したアルゴリズムしか
利用できない。
[ノート ]
IKEv2 では、  IKEv1のハッシュアルゴリズムに相当する折衝パラメーターとして、認証アルゴリズム  (Integrity
Algorithm) と PRF(Pseudo-Random Function) がある。 IKEv2 で ipsec ike proposal-limitation  コマンドが  on に設定され
ているとき、本コマンドで設定されたアルゴリズムを提案する。 ipsec ike proposal-limitation  コマンドが  off に設定
されているとき、または、 ipsec ike proposal-limitation  コマンドに対応していない機種では、本コマンドの設定にか
かわらず、サポートするすべてのアルゴリズムを同時に提案し、相手側セキュリティ・ゲートウェイに選択させる。
また応答側として働く場合は、提案されたものからより安全なアルゴリズムを選択する。
IKEv2 でサポート可能な認証アルゴリズム及び応答時の選択の優先順位は以下の通り。
• HMAC-SHA2-512-256 ＞ HMAC-SHA2-384-192 ＞ HMAC-SHA2-256-128 ＞ HMAC-SHA-1-96 ＞ HMAC-MD5-96
※HMAC-SHA2-512-256 、HMAC-SHA2-384-192 は、 vRX Amazon EC2 版、 vRX VMware ESXi 版、 RTX5000 、
RTX3500 、RTX1220 、RTX1210 、RTX840、および  RTX830 では使用不可能。
また、  IKEv2 でサポート可能な  PRF 、及び応答選択時の優先順位は以下の通り。
• HMAC-SHA2-512 ＞ HMAC-SHA2-384 ＞ HMAC-SHA2-256 ＞ HMAC-SHA-1 ＞ HMAC-MD5
※HMAC-SHA2-512 、HMAC-SHA2-384 は、 vRX Amazon EC2 版、 vRX VMware ESXi 版、 RTX5000 、RTX3500 、
RTX1220 、RTX1210 、RTX840、および  RTX830 では使用不可能。
algorithm  パラメータの  sha384 および  sha512 キーワードは、 vRX Amazon EC2 版、 vRX VMware ESXi 版、 RTX5000 、
RTX3500 、RTX1220 、RTX1210 、RTX840、および  RTX830 では指定不可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id304 | コマンドリファレンス  | IPsec の設定

ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.26 受信したパケットの  SPI 値が無効な値の場合にログに出力するか否かの設定
[書式 ]
ipsec  log illegal-spi  switch
no ipsec  log illegal-spi
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on ログに出力する
off ログに出力しない
• [初期値 ] : off
[説明 ]
IPsec で、受信したパケットの  SPI 値が無効な値の場合に、その旨をログに出力するか否かを設定する。 SPI 値と相
手の  IP アドレスがログに出力される。
無効な  SPI 値を含むパケットを大量に送り付けられることによる  DoS の可能性を減らすため、ログは  1 秒あたり最
大 10 種類のパケットだけを記録する。実際に受信したパケットの数を知ることはできない。
[ノート ]
鍵交換時には、鍵の生成速度の差により一方が新しい鍵を使い始めても他方ではまだその鍵が使用できない状態に
なっているためにこのログが一時的に出力されてしまうことがある。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.27 IKE ペイロードのタイプの設定
[書式 ]
ipsec  ike payload  type gateway_id  type1  [type2 ]
no ipsec  ike payload  type gateway_id  [type1  ...]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•type1  : IKEv1 のメッセージのフォーマット
• [設定値 ] :
設定値 説明
1 ヤマハルーターのリリース  2 との互換性を保持する
2 ヤマハルーターのリリース  3 に合わせるコマンドリファレンス  | IPsec の設定  | 305

設定値 説明
3 初期ベクトル  (IV) の生成方法を一部の実装に合わせ
る
• [初期値 ] : 2
•type2  : IKEv2 のメッセージのフォーマット
• [設定値 ] :
設定値 説明
1 ヤマハルーターの  IKEv2 のリリース  1 との互換性を
保持する
2 鍵交換や鍵の使用方法を一部の実装に合わせる
• [初期値 ] : 2
[説明 ]
IKEv1 および  IKEv2 のペイロードのタイプを設定する。
IKEv1 でヤマハルーターの古いリビジョンと接続する場合には、 type1  パラメータを  1 に設定する必要がある。
IKEv2 でヤマハルーターの以下のリビジョンと接続する場合には、 type2  パラメータを  1 に設定する必要がある。
機種 リビジョン
FWX120 Rev.11.03.02
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.28 IKEv1 鍵交換タイプの設定
[書式 ]
ipsec  ike backward-compatibility  gateway_id  type
no ipsec  ike backward-compatibility  gateway_id  [type]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•type : IKEv1 で使用する鍵交換のタイプ
• [設定値 ] :
設定値 説明
1 ヤマハルーターのリリース  1 (過去のリリース ) との
互換性を保持する
2 ヤマハルーターのリリース  2 (新リリース ) に合わせ
る306 | コマンドリファレンス  | IPsec の設定

• [初期値 ] : 1
[説明 ]
IKEv1 で使用する鍵交換のタイプを設定する。
IKEv1 でヤマハルーターの古いリビジョンと接続する場合には、 type パラメータを  1 に設定する必要がある。
[ノート ]
RTX5000 は Rev.14.00.12 以降で使用可能。
RTX3500 は Rev.14.00.12 以降で使用可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.29 IKE の情報ペイロードを送信するか否かの設定
[書式 ]
ipsec  ike send  info gateway_id  info
no ipsec  ike send  info gateway_id  [info]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•info
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
[説明 ]
IKEv1 動作時に、情報ペイロードを送信するか否かを設定する。受信に関しては、この設定に関わらず、すべての
情報ペイロードを解釈する。
[ノート ]
このコマンドは、接続性の検証などの特別な目的で使用される。定常の運用時は  on に設定する必要がある。
本コマンドは  IKEv2 としての動作には影響を与えない。  IKEv2 では常に、必要に応じて情報ペイロードの送受信を
行う。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_idコマンドリファレンス  | IPsec の設定  | 307

ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.30 PFS を用いるか否かの設定
[書式 ]
ipsec  ike pfs gateway_id  pfs
no ipsec  ike pfs gateway_id  [pfs]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•pfs
• [設定値 ] :
設定値 説明
on 用いる
off 用いない
• [初期値 ] : off
[説明 ]
IKE の始動側として働く場合に、  PFS(Perfect Forward Secrecy) を用いるか否かを設定する。応答側として働く場合
は、このコマンドの設定に関係なく、相手側セキュリティ･ゲートウェイの PFSの使用有無に合わせて動作する。
ただし、 IKEv1 として動作し、且つ  ipsec ike negotiate-strictly  コマンドが  on の場合は、本コマンドの設定と相手側
セキュリティ･ゲートウェイの  PFS の使用有無が一致していなければならない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.31 XAUTH の設定
[書式 ]
ipsec  ike xauth  myname  gateway_id  name  password
no ipsec  ike xauth  myname  gateway_id308 | コマンドリファレンス  | IPsec の設定

[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•name
• [設定値 ] : XAUTH で通知する名前  (32 文字以内  )
• [初期値 ] : -
•password
• [設定値 ] : XAUTH で通知するパスワード  (32 文字以内  )
• [初期値 ] : -
[説明 ]
XAUTH の認証を要求されたときに通知する名前とパスワードを設定する。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.32 XAUTH 認証、 EAP-MD5 認証に使用するユーザ  ID の設定
[書式 ]
auth  user userid  username  password
no auth  user userid  [username  ...]
[設定値及び初期値 ]
•userid
• [設定値 ] :
•ユーザ識別番号
設定値 機種
1..6000vRX Amazon EC2 版 ( 通常モード  での動作時の
み )、vRX VMware ESXi 版 ( 通常モード  での動作時
のみ  )、vRX さくらのクラウド版
1..3000 RTX5000
1..1000 その他
• [初期値 ] : -
•username
• [設定値 ] :
•ユーザ名  ( 256 文字以内、 3 文字以上を設定してください  )
• [初期値 ] : -
•password
• [設定値 ] :
•パスワード  ( 64 文字以内、 3 文字以上を設定してください  )
• [初期値 ] : -コマンドリファレンス  | IPsec の設定  | 309

[説明 ]
IKEv1 の XAUTH 認証、または  IKEv2 の EAP-MD5 認証に使用するユーザ  ID を設定する。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•userid
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX2 1000 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.33 XAUTH 認証、 EAP-MD5 認証に使用するユーザ  ID の属性の設定
[書式 ]
auth  user attribute  userid  attribute =value  [attribute =value  ...]
no auth  user attribute  userid  [attribute =value  ...]
[設定値及び初期値 ]
•userid
• [設定値 ] :
•ユーザ識別番号
設定値 機種
1..6000vRX Amazon EC2 版 ( 通常モード  での動作時の
み )、vRX VMware ESXi 版 ( 通常モード  での動作時
のみ  )、vRX さくらのクラウド版
1..3000 RTX5000
1..1000 その他
• [初期値 ] : -
•attribute=value
• [設定値 ] : ユーザ属性
• [初期値 ] : xauth=off
[説明 ]
IKEv1 の XAUTH 認証、または  IKEv2 の EAP-MD5 認証に使用するユーザ  ID の属性を設定する。
設定できる属性は以下のとおり。
attribute value 説明
xauthonIPsec の XAUTH 認証にこの  ID を使
用する
offIPsec の XAUTH 認証にこの  ID を使
用しない
xauth-addressIP address[/netmask](IPv6 アドレス
可 )IPsec の接続時に、このアドレスを内
部 IP アドレスとして通知する
xauth-dns IP address(IPv6 アドレス可  )IPsec の接続時に、このアドレスを
DNS サーバーアドレスとして通知す
る310 | コマンドリファレンス  | IPsec の設定

attribute value 説明
xauth-wins IP address(IPv6 アドレス可  )IPsec の接続時に、このアドレスを
WINS サーバーアドレスとして通知
する
xauth-filter フィルタセットの名前を表す文字列IPsec の接続時に、このフィルタを適
用する
eap-md5onIKEv2 の EAP-MD5 認証にこの  ID を
使用する
offIKEv2 の EAP-MD5 認証にこの  ID を
使用しない
同じ属性が重複して指定されている場合はコマンドエラーとなる。
[ノート ]
本コマンドにて明示的に設定した属性値は、該当のユーザ  ID が属しているユーザグループに対して、 auth user
group attribute  コマンドによって設定された属性値に優先して適用される。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•userid
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX2 1000 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.34 XAUTH 認証、 EAP-MD5 認証に使用するユーザグループの設定
[書式 ]
auth  user group  groupid  userid  [userid  ...]
no auth  user group  groupid
[設定値及び初期値 ]
•groupid
• [設定値 ] :
•ユーザグループ識別番号
設定値 機種
1..6000vRX Amazon EC2 版 ( 通常モード  での動作時の
み )、vRX VMware ESXi 版 ( 通常モード  での動作時
のみ  )、vRX さくらのクラウド版
1..3000 RTX5000
1..1000 その他
• [初期値 ] : -
•userid
• [設定値 ] : ユーザ識別番号もしくはユーザ識別番号の範囲  ( 複数指定することが可能  )
• [初期値 ] : -
[説明 ]
IKEv1 の XAUTH 認証、または  IKEv2 の EAP-MD5 認証に使用するユーザグループを設定する。コマンドリファレンス  | IPsec の設定  | 311

[設定例 ]
# auth user group 1 100 101 102
# auth user group 1 200-300
# auth user group 1 100 103 105 107-110 113
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•groupid
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX2 1000 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
•userid
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX2 1000 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.35 XAUTH 認証、 EAP-MD5 認証に使用するユーザグループの属性の設定
[書式 ]
auth  user group  attribute  groupid  attribute =value  [attribute =value  ...]
no auth  user group  attribute  groupid  [attribute =value  ...]
[設定値及び初期値 ]
•groupid
• [設定値 ] :
•ユーザグループ識別番号
設定値 機種
1..6000vRX Amazon EC2 版 ( 通常モード  での動作時の
み )、vRX VMware ESXi 版 ( 通常モード  での動作時
のみ  )、vRX さくらのクラウド版
1..3000 RTX5000 、RTX3510
1..1000 その他
• [初期値 ] : -
•attribute=value
• [設定値 ] : ユーザグループ属性
• [初期値 ] : xauth=off
[説明 ]
IKEv1 の XAUTH 認証、または  IKEv2 の EAP-MD5 認証に使用するユーザグループの属性を設定する。
設定できる属性は以下のとおり。312 | コマンドリファレンス  | IPsec の設定

attribute value 説明
xauthonIPsec の XAUTH 認証にこのグループ
に含まれるユーザ  ID を使用する
offIPsec の XAUTH 認証にこのグループ
に含まれるユーザ  ID を使用しない
xauth-address-poolIP アドレスの範囲  (IPv6 アドレス
可 )IPsec の接続時に、このアドレスプー
ルからアドレスを選択し、内部  IP ア
ドレスとして通知する
xauth-dns IP address(IPv6 アドレス可  )IPsec の接続時に、このアドレスを
DNS サーバーアドレスとして通知す
る
xauth-wins IP address(IPv6 アドレス可  )IPsec の接続時に、このアドレスを
WINS サーバーアドレスとして通知
する
xauth-filter フィルタセットの名前を表す文字列IPsec の接続時に、このフィルタを適
用する
eap-md5onIKEv2 の EAP-MD5 認証にこの  ID を
使用する
offIKEv2 の EAP-MD5 認証にこの  ID を
使用しない
xauth-address-pool の属性値である  IP アドレスの範囲は、以下のいずれかの書式にて記述する。
• IP address[/netmask]
• IP address-IP address[/netmask]
同じ属性が重複して指定されている場合はコマンドエラーとなる。
[ノート ]
本コマンドで設定した属性値は、該当のユーザグループに含まれるすべてのユーザに対して有効となる。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•groupid
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX2 1000 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.36 XAUTH によるユーザ認証の設定
[書式 ]
ipsec  ike xauth  request  gateway_id  auth [group_id ]
no ipsec  ike xauth  request  gateway_id  [auth ...]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティゲートウェイの識別子
• [初期値 ] : -
•group_id
• [設定値 ] : .認証に使用するユーザグループの識別番号
• [初期値 ] : -コマンドリファレンス  | IPsec の設定  | 313

•auth
• [設定値 ] :
設定値 説明
on 要求する
off 要求しない
• [初期値 ] : off
[説明 ]
IPsec の認証を行う際、 Phase1 終了後に  XAUTH によるユーザ認証をクライアントに要求するか否かを設定する。
group_id  を指定した場合には、該当のユーザグループに含まれるユーザを認証の対象とする。
group_id  の指定がない場合や、指定したユーザグループに含まれるユーザ情報では認証できなかった場合、 RADIUS
サーバーの設定があれば  RADIUS サーバーを用いた認証を追加で試みる。
[ノート ]
本コマンドによる設定はルーターが受動側として動作する時にのみ有効であり、始動側のセキュリティゲートウェ
イから送信された  isakmp SA パラメータの提案に、認証方式として  XAUTHInitPreShared(65001) が含まれていた場合
に、この提案を受け入れ、 XAUTH によるユーザ認証を行う。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
•groupid
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX2 1000 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.37 内部  IP アドレスプールの設定
[書式 ]
ipsec  ike mode-cfg  address  pool pool_id  ip_address [/mask ]
ipsec  ike mode-cfg  address  pool pool_id  ip_address-ip_address [/mask ]
no ipsec  ike mode-cfg  address  pool pool_id  [ip_address  ...]
[設定値及び初期値 ]
•pool_id
• [設定値 ] : アドレスプール  ID(1..65535)
• [初期値 ] : -
•ip_address
• [設定値 ] : IP アドレス  (IPv6 アドレス可  )
• [初期値 ] : -314 | コマンドリファレンス  | IPsec の設定

•ip_address-ip_address
• [設定値 ] : IP アドレスの範囲  (IPv6 アドレス可  )
• [初期値 ] : -
•mask
• [設定値 ] : ネットマスク  (IPv6 アドレスの時はプレフィックス長  )
• [初期値 ] : -
[説明 ]
IPsec クライアントに割り当てる内部  IP アドレスのアドレスプールを設定する。
本コマンドにて設定したアドレスプールは、 ipsec ike mode-cfg address  gateway_id  ...コマンドにて用いられる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.38 IKE XAUTH Mode-Cfg メソッドの設定
[書式 ]
ipsec  ike mode-cfg  method  gateway_id  method  [option ]
no ipsec  ike mode-cfg  method  gateway_id  [method ...]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティゲートウェイの識別子
• [初期値 ] : -
•method
• [設定値 ] :
設定値 説明
set SET メソッド
• [初期値 ] : set
•option
• [設定値 ] :
設定値 説明
openswan Openswan 互換モード
• [初期値 ] : -
[説明 ]
IKE XAUTH の Mode-Cfg でのアドレス割り当てメソッドを設定する。指定できるのは  SET メソッドのみである。
option  に 'openswan' を指定した場合には  Openswan 互換モードとなり、 Openswan と接続できるようになる。
[ノート ]
ダイヤルアップ  VPN の発呼側にヤマハルーター  を利用するときに、 option  を指定していると  XAUTH では接続でき
ない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100コマンドリファレンス  | IPsec の設定  | 315

ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.39 IPsec クライアントに割り当てる内部  IP アドレスプールの設定
[書式 ]
ipsec  ike mode-cfg  address  gateway_id  pool_id
no ipsec  ike mode-cfg  address  gateway_id  [pool_id ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティゲートウェイの識別子
• [初期値 ] : -
•pool_id
• [設定値 ] : アドレスプール  ID
• [初期値 ] : -
[説明 ]
IPsec クライアントに内部  IP アドレスを割り当てる際に参照する、内部  IP アドレスプールを設定する。
内部  IP アドレスの  IPsec クライアントへの通知は、 XAUTH 認証に使用する  Config-Mode にて行われるため、 XAUTH
認証を行わない場合には通知は行われない。
以下のいずれかの方法にて、認証ユーザ毎に割り当てる内部  IP アドレスが設定されている場合には、アドレスプー
ルからではなく、個別に設定されているアドレスを通知する。
• RADIUS サーバーに登録されている場合
•以下のコマンドを用いて設定されている場合
•auth user attribute  userid  xauth-address= address [/mask ]
•auth user group attribute  groupid  xauth-address-pool= address-address [/mask ]
アドレスプールに登録されているアドレスが枯渇した場合には、アドレスの割当を行わない。
[ノート ]
VPN クライアントとして  YMS-VPN1 を用いる場合、 XAUTH 認証を行うためには必ず内部  IP アドレスの通知を行
う設定にしなければならない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830316 | コマンドリファレンス  | IPsec の設定

15.40 VPN クライアントの同時接続制限ライセンスの登録
[書式 ]
ipsec  ike license-key  license_id  key
no ipsec  ike license-key  license_id  [...]
[設定値及び初期値 ]
•license_id
• [設定値 ] : ルーターキーの識別番号  (1..500)
• [初期値 ] : -
•key
• [設定値 ] : ルーターキー  (64文字以内 )
• [初期値 ] : -
[説明 ]
VPN クライアントソフト  (同時接続版 ) からの VPN接続を受け入れるためのルーターキー  (ライセンスキー ) を設定
する。
各ルーターキーには固有の同時接続数が付与されており、異なる複数のルーターキーを登録することで、各ルータ
ーキーの合計分の最大同時接続数を確保することができる。このとき、 VPN クライアントソフトは本コマンドで登
録したルーターキーに対応するクライアントキーならばどれを使用してもよい。 VPN クライアントソフトが使用す
るクライアントキーに関わらず、登録された各ルーターキーの合計の最大同時接続数を基に接続制限が施される。
[ノート ]
RTX830 は Rev.15.02.15 以降で使用可能。
[設定例 ]
[YMS-VPN1-CP/YMS-VPN7-CP の場合 ]
# tunnel select 1
#  tunnel template 2-20
#  ipsec tunnel 1
#   ipsec sa policy 1 1 esp aes-cbc sha-hmac
#   ipsec ike log 1 payload-info
#   ipsec ike remote address 1 any
#   ipsec ike xauth request 1 on 11
#   ipsec ike mode-cfg address 1 1
#   ipsec ike license-key use 1 on
#  tunnel enable 1
# ipsec ike license-key 1 abcdefg-10-hijklmno
# ipsec ike license-key 2 pqrstuv-10-wxyz0123
# ipsec ike mode-cfg address pool 1 172.16.0.1-172.16.0.20/32
# auth user 1 user1 pass1
# auth user 2 user2 pass2
   :
# auth user 20 user20 pass20
# auth user group 11 1-20
# auth user group attribute 11 xauth=on xauth-dns=10.10.10.1
[YMS-VPN8-CP の場合 ]
# pp select anonymous
#  pp bind tunnel1-tunnel20
#  pp auth request mschap-v2
#  pp auth username user1 pass1
#  pp auth username user2 pass2
   :
#  pp auth username user20 pass20
#  ppp ipcp ipaddress on
#  ppp ipcp msext on
#  ip pp remote address pool 172.16.0.1-172.16.0.20
#  ip pp mtu 1258
#  pp enable anonymous
# tunnel select 1
#  tunnel encapsulation l2tp
#  ipsec tunnel 1
#   ipsec sa policy 1 1 esp 3des-cbc sha-hmac
#   ipsec ike keepalive use 1 offコマンドリファレンス  | IPsec の設定  | 317

#   ipsec ike local address 1 172.16.0.254
#   ipsec ike remote address 1 any
#   ipsec ike license-key use 1 on
#   l2tp tunnel disconnect time off
#   ip tunnel tcp mss limit auto
#  tunnel enable 1
   :
# ipsec ike license-key 1 abcdefg-10-hijklmno
# ipsec ike license-key 2 pqrstuv-10-wxyz0123
# ipsec transport 1 1 udp 1701
# ipsec auto refresh on
# l2tp service on
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.41 VPN クライアントの同時接続制限ライセンスの適用
[書式 ]
ipsec  ike license-key  use gateway_id  sw
no ipsec  ike license-key  use gateway_id  [...]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•sw
• [設定値 ] :
設定値 説明
on ルーターキーの適用を許可する
off ルーターキーの適用を許可しない
• [初期値 ] : off
[説明 ]
VPN クライアントソフト  (同時接続版 ) からの VPN接続を受け入れるためのルーターキー  ( ライセンスキー  ) の適
用を許可するか否かを設定する。
ルーターキーの適用を許可されたゲートウェイが、対応するクライアントキーを持つ  VPN クライアントソフトと接
続可能になる。
[ノート ]
RTX830 は Rev.15.02.15 以降で使用可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830318 | コマンドリファレンス  | IPsec の設定

15.42 IKE のログの種類の設定
[書式 ]
ipsec  ike log [gateway_id ] type [type]
no ipsec  ike log [gateway_id ] [type]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•type
• [設定値 ] :
設定値 説明
message-info IKE メッセージの内容
payload-info ペイロードの処理内容
key-info 鍵計算の処理内容
• [初期値 ] : -
[説明 ]
出力するログの種類を設定する。ログはすべて、 debug レベルの  SYSLOG で出力される。
IKEv2に対応した機種では、 gateway_id  パラメータを省略することができる。  gateway_id  パラメータを省略した設
定は、応答側として働く際、セキュリティ・ゲートウェイが特定できない時点での通信に対して適用される。
[ノート ]
このコマンドが設定されていない場合には、最小限のログしか出力しない。複数の type パラメータを設定すること
もできる。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.43 ESP を UDP でカプセル化して送受信するか否かの設定
[書式 ]
ipsec  ike esp-encapsulation  gateway_id  encap
no ipsec  ike esp-encapsulation  gateway_id
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•encap
• [設定値 ] :コマンドリファレンス  | IPsec の設定  | 319

設定値 説明
on ESP を UDP でカプセル化して送信する
off ESP を UDP でカプセル化しないで送信する
• [初期値 ] : off
[説明 ]
NAT などの影響で  ESP が通過できない環境で  IPsec の通信を確立するために、 ESP を UDP でカプセル化して送受信
できるようにする。このコマンドの設定は双方のルーターで一致させる必要がある。
[ノート ]
ipsec ike nat-traversal  コマンドとの併用はできない。
本コマンドは  IKEv2 により確立された  SA を伴う  IPsec 通信には影響を与えない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.44 折衝パラメーターを制限するか否かの設定
[書式 ]
ipsec  ike proposal-limitation  gateway_id  switch
no ipsec  ike proposal-limitation  gateway_id  [switch ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on 折衝パラメーターを制限する
off 折衝パラメーターを制限しない
• [初期値 ] : off
[説明 ]
IKEv2 で鍵交換を始動するときに、 SA を構築するための各折衝パラメーターを、特定のコマンド設定値に限定して
提案するか否かを設定する。このコマンドの設定が  off のときは、サポート可能な折衝パラメーター全てを提案す
る。
このコマンドが適用されるパラメーターと対応するコマンドは以下の通りである。
パラメーター コマンド
暗号アルゴリズム ipsec ike encryption320 | コマンドリファレンス  | IPsec の設定

パラメーター コマンド
グループ ipsec ike group
ハッシュアルゴリズム ipsec ike hash
暗号・認証アルゴリズム ipsec sa policy  ※CHILD SA 作成時
[ノート ]
本コマンドは  IKEv2 でのみ有効であり、 IKEv1 の動作に影響を与えない。
RTX1210 は Rev.14.01.09 以降で使用可能。
RTX5000 、RTX3500 は Rev.14.00.18 以降で使用可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.45 IKE のメッセージ  ID 管理の設定
[書式 ]
ipsec  ike message-id-control  gateway_id  switch
no ipsec  ike message-id-control  gateway_id  [switch ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on リクエストメッセージの送信をメッセージ  ID で管理
する
off リクエストメッセージの送信をメッセージ  ID で管理
しない
• [初期値 ] : off
[説明 ]
自機から  IKEv2 のリクエストメッセージを送信するときのメッセージ  ID 管理方法を設定する。
on に設定しているとき、同じ  IKE SA を使用して送信済みの  IKE メッセージに対する全てのレスポンスメッセージ
を受信していない場合、新しい  IKE メッセージは送信しない。
[ノート ]
本コマンドは  IKEv2 でのみ有効であり、 IKEv1 の動作に影響を与えない。コマンドリファレンス  | IPsec の設定  | 321

RTX1210 は Rev.14.01.09 以降で使用可能。
RTX5000 、RTX3500 は Rev.14.00.18 以降で使用可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.46 CHILD SA 作成方法の設定
[書式 ]
ipsec  ike child-exchange  type gateway_id  type
no ipsec  ike child-exchange  type gateway_id  [type]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•type : IKEv2 の CHILD SA 作成方法のタイプ
• [設定値 ] :
設定値 説明
1 ヤマハルーターの  IKEv2 の従来の動作との互換性を
保持する
•鍵更新の失敗などにより有効な  Child SA がなく
なったとき、 IKE_AUTH 交換を利用して新しく
Child SA を作成する。
• CREATE_CHILD_SA 交換を利用して  IKE SA の鍵
更新を行う。ただし、 CREATE_CHILD_SA 交換で
の鍵更新が指定された回数実施されると次回の鍵
更新のタイミングで  IKE_SA_INIT 交換により
IKE SA を作成する。回数の指定は  ipsec ike retry
コマンドの  count  パラメータで行う。
2 CREATE_CHILD_SA 交換を一部の実装にあわせる
•鍵更新の失敗などにより有効な  Child SA がなく
なったとき、 CREATE_CHILD_SA 交換を利用して
新しく  Child SA を作成する。
•常に  CREATE_CHILD_SA 交換を利用して  IKE SA
の鍵更新を行う。
• [初期値 ] : 1
[説明 ]
IKEv2 の CHILD SA 作成方法を設定する。
このコマンドに対応する機種同士で接続する場合、 typeを同じ設定にして接続する必要がある。322 | コマンドリファレンス  | IPsec の設定

[ノート ]
本コマンドは  IKEv2 でのみ有効であり、 IKEv1 の動作に影響を与えない。
RTX1210 は Rev.14.01.11 以降で使用可能。
RTX5000 、RTX3500 は Rev.14.00.18 以降で使用可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.47 鍵交換の始動パケットを受信するか否かの設定
[書式 ]
ipsec  ike negotiation  receive  gateway_id  switch
no ipsec  ike negotiation  receive  gateway_id  [switch ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on 鍵交換の始動パケットを受信する
off 鍵交換の始動パケットを受信しない
• [初期値 ] : on
[説明 ]
IKEv2で、鍵交換の始動パケットを受信するか否かを設定する。
受信しないに設定した場合は、結果として受動側としては動作せず、必ず始動側として動作するようになる。
[ノート ]
本コマンドは  IKEv2 でのみ有効であり、 IKEv1 の動作に影響を与えない。
offにする場合には、 ipsec ike remote address または ipsec ike remote name をIPアドレスで設定しておく必要がある。
RTX5000 Rev.14.00.26 以降、 RTX3500 Rev.14.00.26 以降、 RTX1210 Rev.14.01.26 以降、 RTX830 Rev.15.02.03 以降  で使
用可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_idコマンドリファレンス  | IPsec の設定  | 323

ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.48 SA 関連の設定
再起動されるとすべての  SA がクリアされることに注意しなくてはいけない。
15.48.1 SA の寿命の設定
[書式 ]
ipsec  ike duration  sa gateway_id  second  [kbytes ] [rekey rekey ] [forced-reduction= del_time ]
no ipsec  ike duration  sa gateway_id  [second  [kbytes ] [rekey rekey ] [forced-reduction= del_time ]]
[設定値及び初期値 ]
•sa
• [設定値 ] :
設定値 説明
ipsec-sa ( もしくは  child-sa) IPsec SA (CHILD SA)
isakmp-sa ( もしくは  ike-sa) ISAKMP SA (IKE SA)
• [初期値 ] : -
•gateway_id
• [設定値 ] : セキュリティー・ゲートウェイの識別子
• [初期値 ] : -
•second
• [設定値 ] : 秒数  (300..691200)
• [初期値 ] : 28800 秒
•kbytes
• [設定値 ] :
設定値 説明
100..100000キロ単位のバイト数  ( RTX5000 / RTX3500
Rev.14.00.33 以前、 RTX1220 Rev.15.04.04 以前、
RTX1210 Rev.14.01.41 以前、 RTX830 Rev.15.02.26 以
前 )
100..2147483647 キロ単位のバイト数  ( 上記以外の機種・リビジョン  )
offバイト寿命を設定しない  ( RTX5000 / RTX3500
Rev.14.00.33 以前、 RTX1220 Rev.15.04.04 以前、
RTX1210 Rev.14.01.41 以前、および  RTX830
Rev.15.02.26 以前では指定不可能  )
• [初期値 ] :
• - ( RTX5000 / RTX3500 Rev.14.00.33 以前、 RTX1220 Rev.15.04.04 以前、 RTX1210 Rev.14.01.41 以前、 RTX830
Rev.15.02.26 以前  )
• 2000000 ( 上記以外の機種・リビジョン  )
•rekey  : SA を更新するタイミング
• [設定値 ] :324 | コマンドリファレンス  | IPsec の設定

設定値 説明
70%-90% パーセント
off更新しない  (sa パラメーターで  isakmp-sa (ike-sa) を
指定したときのみ設定可能  )
• [初期値 ] : 75%
•del_time
• [設定値 ] : 古くなった SAの寿命を強制的に短縮する時間 (1..691200)
• [初期値 ] : -
[説明 ]
各 SA の寿命を設定する。
kbytes  パラメーターを指定した場合には、  second  パラメーターで指定した時間が経過するか、指定したバイト数の
データを処理した後に  SA は消滅する。  kbytes  パラメーターは  sa パラメーターとして  ipsec-sa (child-sa) を指定した
ときのみ有効である。 SA の更新は  kbytes  パラメーターに設定したバイト数の  75％ を処理したタイミングで行われ
る。また、 IPsec SA (CHILD SA) が更新されたとき、古くなった既存の  IPsec SA (CHILD SA) の寿命が  30 秒以上であ
る場合は、寿命が  30 秒に短縮される。
rekey  パラメーターは  SA を更新するタイミングを決定する。例えば、  second  パラメーターで  20000 を指定し、  rekey
パラメーターで 75%を指定した場合には、  SA を生成してから  15000 秒経過したときに新しい  SA を生成する。
rekey  パラメーターは  second  パラメーターに対する比率を表すもので、  kbytes  パラメーターの値とは関係がない。
sa パラメーターで  isakmp-sa (ike-sa) を指定したときに限り、 rekey  パラメーターで  'off' を設定できる。このとき、
IPsec SA (CHILD SA) を作る必要がない限り、  ISAKMP SA (IKE SA) の更新を保留するので、  ISAKMP SA (IKE SA)
の生成を最小限に抑えることができる。
その他、動作する IKEのバージョンによって異なる、本コマンドの影響、注意点については以下の通り。
• IKEv1
始動側として働く場合に、このコマンドで設定した寿命値が提案される。応答側として働く場合は、このコマン
ドの設定に関係なく相手側から提案された寿命値に合わせる。
また、  ISAKMP SA に対する  rekey  パラメーターを  off に設定した場合、その効果を得るためには、次の 2点に注
意して設定する必要がある。
1.IPsec SA よりも  ISAKMP SA の寿命を短く設定する。
2.ダングリング  SA を許可する。すなわち、  ipsec ike restrict-dangling-sa  コマンドの設定を  off にする。
RTX5000 / RTX3500 Rev.14.00.33 以前、 RTX1220 Rev.15.04.04 以前、 RTX1210 Rev.14.01.41 以前、 RTX830
Rev.15.02.26 以前、または  NVR700W Rev.15.00.23 以前と接続する場合は、 kbytes  パラメーターを  2 GB 以下に設
定する必要がある。これらの機種・リビジョンは、 2 GB を超えるバイト寿命値を正しく認識できない。
• IKEv2
IKEv2 では  SA 寿命値は折衝されず、各セキュリティー・ゲートウェイが独立して管理するものとなっている。
従って、確立された  SA には、常にこのコマンドで設定した寿命値がセットされる。ただし、相手側セキュリテ
ィー・ゲートウェイの方が  SA 更新のタイミングが早ければ、  SA はその分早く更新されることになる。
forced-reduction オプションに時間を指定すると、 SA を更新した際に古くなった既存の  SA の寿命を強制的に設定値
に変更し、消滅までの時間を早めることができる。ただし、 IPsec SA (CHILD SA) で kbytes  パラメーターにバイト寿
命値を指定している場合は、 del_time  パラメーターで  31 秒以上の値を設定していても、短縮される値は  30 秒とな
る。また、 IKEv1 では寿命が設定値よりも短い場合は変更しない。
ISAKMP SA (IKE SA) の寿命が  IPsec SA (CHILD SA) の寿命より先に尽きた場合は、 ISAKMP SA (IKE SA) の寿命値
を IPsec SA (CHILD SA) の寿命値に合わせる。
なお、このコマンドを設定しても、すでに存在する  SA の寿命値は変化せず、新しく作られる  SA にのみ、新しい寿
命値が適用される。コマンドリファレンス  | IPsec の設定  | 325

[ノート ]
forced-reduction オプションは、 vRX Amazon EC2 版、 vRX VMware ESXi 版、 RTX5000 / RTX3500 Rev.14.00.33 以前、
RTX1300 Rev.23.00.04 以前、 RTX1220 Rev.15.04.04 以前、 RTX1210 Rev.14.01.41 以前、および  RTX830 Rev.15.02.26 以
前では指定できない。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.48.2 SA のポリシーの定義
[書式 ]
ipsec  sa policy  policy_id  gateway_id  ah [ah_algorithm ] [local-id= local-id ] [remote-id= remote-id ] [anti-replay-
check= check ]
ipsec  sa policy  policy_id  gateway_id  esp [ esp_algorithm ] [ah_algorithm ] [anti-replay-check= check ]
no ipsec  sa policy  policy_id  [gateway_id ]
[設定値及び初期値 ]
•policy_id
• [設定値 ] : ポリシー ID(1..2147483647)
• [初期値 ] : -
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
• ah : 認証ヘッダ  (Authentication Header) プロトコルを示すキーワード
• [初期値 ] : -
• esp : 暗号ペイロード  (Encapsulating Security Payload) プロトコルを示すキーワード
• [初期値 ] : -
•ah_algorithm  : 認証アルゴリズム
• [設定値 ] :
設定値 説明
md5-hmac HMAC-MD5
sha-hmac HMAC-SHA-1
sha256-hmac HMAC-SHA2-256
sha384-hmac HMAC-SHA2-384
sha512-hmac HMAC-SHA2-512
• [初期値 ] :
• sha-hmac ( AH プロトコルの場合  )
• - ( ESP プロトコルの場合  )
•esp_algorithm  : 暗号アルゴリズム
• [設定値 ] :326 | コマンドリファレンス  | IPsec の設定

設定値 説明
3des-cbc 3DES-CBC
des-cbc DES-CBC
aes-cbc AES128-CBC
aes256-cbc AES256-CBC
• [初期値 ] : aes-cbc
•local-id
• [設定値 ] : 自分側のプライベートネットワーク
• [初期値 ] : -
•remote-id
• [設定値 ] : 相手側のプライベートネットワーク
• [初期値 ] : -
•check
• [設定値 ] :
設定値 説明
on シーケンス番号のチェックを行う
off シーケンス番号のチェックを行わない
• [初期値 ] :
• off (RTX3510 、RTX1300)
• on (上記以外 )
[説明 ]
SA のポリシーを定義する。この定義はトンネルモードおよびトランスポートモードの設定に必要である。この定
義は複数のトンネルモードおよびトランスポートモードで使用できる。
local-id  、 remote-id  には、カプセル化したいパケットの始点／終点アドレスの範囲をネットワークアドレスで記述す
る。これにより、  1 つのセキュリティ・ゲートウェイに対して、複数の  IPsec SA を生成し、  IP パケットの内容に応
じて  SA を使い分けることができるようになる。
check =on の場合、受信パケット毎にシーケンス番号の重複や番号順のチェックを行い、エラーとなるパケットは破
棄する。破棄する際には  debug レベルで
[IPSEC] sequence difference
[IPSEC] sequence number is wrong
といったログが記録される。
相手側が、トンネルインタフェースでの優先 /帯域制御を行っている場合、シーケンス番号の順序が入れ替わってパ
ケットを受信することがある。その場合、実際にはエラーではないのに上のログが表示され、パケットが破棄され
ることがあるので、そのような場合には設定を  off にするとよい。
IKEv2 では、 ipsec ike proposal-limitation  コマンドが  on に設定されているとき、本コマンドの  ah_algorithm  、および
esp_algorithm  パラメーターで設定されたアルゴリズムを提案する。 ipsec ike proposal-limitation  コマンドが  off に設
定されているとき、または、 ipsec ike proposal-limitation  コマンドに対応していない機種では、本コマンドの設定に
かかわらず、サポートするすべてのアルゴリズムを同時に提案し、相手側セキュリティ・ゲートウェイに選択させ
る。また応答側として働く場合は受け取った提案から以下の優先順位でアルゴリズムを選択する。
•認証アルゴリズム
HMAC-SHA2-512 ＞ HMAC-SHA2-384 ＞ HMAC-SHA2-256 ＞ HMAC-SHA-1 ＞ HMAC-MD5
※HMAC-SHA2-512 、HMAC-SHA2-384 は、 vRX Amazon EC2 版、 vRX VMware ESXi 版、 RTX5000 、RTX3500 、
RTX1220 、RTX1210 、RTX840、および  RTX830 では使用不可。
•暗号アルゴリズム
AES256-CBC ＞ AES192-CBC ＞ AES128-CBC ＞ 3DES-CBC ＞ DES-CBC
※IKEv2 でのみ AES192-CBC をサポートする。ただし、コマンドで AES192-CBC を選択することはできない。
また、  IKEv2 では  local-id  、 remote-id  パラメーターに関しても効力を持たない。コマンドリファレンス  | IPsec の設定  | 327

[ノート ]
双方で設定する local-id  とremote-id  は一致している必要がある。
ah_algorithm  パラメータの  sha384-hmac および  sha512-hmac キーワードは、 vRX Amazon EC2 版、 vRX VMware ESXi
版、 RTX5000 、RTX3500 、RTX1220 、RTX1210 、RTX840、および  RTX830 では指定不可能。
[設定例 ]
# ipsec sa policy 101 1 esp aes-cbc sha-hmac
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.48.3 SA の手動更新
[書式 ]
ipsec  refresh  sa
[説明 ]
SA を手動で更新する。
[ノート ]
管理されている  SA をすべて削除して、 IKE の状態を初期化する。
このコマンドでは、 SA の削除を相手に通知しないので、通常の運用では ipsec sa delete all  コマンドの方が望まし
い。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.48.4 ダングリング  SA の動作の設定
[書式 ]
ipsec  ike restrict-dangling-sa  gateway_id  action
no ipsec  ike restrict-dangling-sa  gateway_id  [action ]
[設定値及び初期値 ]
•gateway_id
• [設定値 ] : セキュリティ・ゲートウェイの識別子
• [初期値 ] : -
•action
• [設定値 ] :
設定値 説明
auto アグレッシブモードの始動側でのみ  IKE SA と IPsec
SA を同期させる
off IKE SA と IPsec SA を同期させない。328 | コマンドリファレンス  | IPsec の設定

• [初期値 ] : auto
[説明 ]
このコマンドは  IKEv1 のダングリング  SA の動作に制限を設ける。
ダングリング  SA とは、 IKE SA を削除するときに対応する  IPsec SA を削除せずに残したときの状態を指す。
RT シリーズでは基本的にはダングリング  SA を許す方針で実装しており、 IKE SA と IPsec SA を独立のタイミング
で削除する。
auto を設定したときには、アグレッシブモードの始動側でダングリング  SA を排除し、 IKE SA と IPsec SA を同期し
て削除する。この動作は  IKE keepalive が正常に動作するために必要な処置である。
off を設定したときには、常にダングリング  SA を許す動作となり、 IKE SA と IPsec SA を独立なタイミングで削除す
る。
ダイヤルアップ  VPN のクライアント側ではない場合には、このコマンドの設定に関わらず常に  IKE SA と IPsec SA
は独立に管理され、削除のタイミングは必ずしも同期しない。
[ノート ]
ダングリング  SA の強制削除が行われても、 通常は新しい  IKE SA に基づいた新しい  IPsec SA が存在するので通信に
支障が出ることはない。
ダイヤルアップ  VPN のクライアント側では、このコマンドにより動作を変更でき、それ以外では、ダングリング  SA
が発生しても何もせず通信を続ける。
ダイヤルアップ  VPN のクライアント側でダングリング  SA を許さないのは、 IKE キープアライブを正しく機能させ
るために必要なことである。
IKE キープアライブでは、 IKE SA に基づいてキープアライブを行う。ダングリング  SA が発生した場合には、その
SA についてはキープアライブを行う  IKE SA が存在せず、キープアライブ動作が行えない。そのため、 IKE キープ
アライブを有効に動作させるにはダングリング  SA が発生したら強制的に削除して、通信は対応する  IKESA が存在
する  IPsec SA で行われるようにしなくてはいけない。
本コマンドは  IKEv2 の動作には影響を与えない。  IKEv2 では仕様として、 ダングリング  SA の存在を禁止している。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.48.5 IPsec NAT トラバーサルを利用するための設定
[書式 ]
ipsec  ike nat-traversal  gateway  switch  [keepalive= interval ] [force= force_switch ] [type= type]
no ipsec  ike nat-traversal  gateway  [switch  ...]
[設定値及び初期値 ]
•gateway
• [設定値 ] : セキュリティゲートウェイの識別子
• [初期値 ] : -
•switch  : 動作の有無
• [設定値 ] :コマンドリファレンス  | IPsec の設定  | 329

設定値 説明
on NAT トラバーサルの動作を有効にする
off NAT トラバーサルの動作を無効にする
• [初期値 ] : off
•interval  : NAT キープアライブの送信間隔
• [設定値 ] :
設定値 説明
off 送信しない
30-100000 時間 [秒]
• [初期値 ] : 300
•force_switch
• [設定値 ] :
設定値 説明
on 通信経路上に  NAT がなくても  NAT トラバーサルを
使用する
off 通信経路上に  NAT がなければ  NAT トラバーサルを
使用しない
• [初期値 ] : off
•type
• [設定値 ] :
設定値 説明
1 ヤマハルーターの従来の動作との互換性を保持する
2 NAT トラバーサル使用時に交換するペイロードを一
部の実装に合わせる
• [初期値 ] : 1
[説明 ]
NAT トラバーサルの動作を設定する。この設定があるときには、 IKE で NAT トラバーサルの交渉を行う。
相手が  NAT トラバーサルに対応していないときや、通信経路上に  NAT の処理がないときには、 NAT トラバーサル
を使用せず、 ESP パケットを使って通信する。
対向のルーターや端末でも  NAT トラバーサルの設定が必要である。いずれか一方にしか設定がないときには、 NAT
トラバーサルを使用せず、 ESP パケットを使って通信する
type に対応した機種同士で接続する場合、 typeを同じ設定にして接続する必要がある。また、 type に 2 を指定した場
合、 type に対応していない機種との接続はできない。
IKEv2 では、イニシエータとして動作する場合のみ  force_switch  パラメータが影響する。このオプションは、通信経
路上に  NAT 処理がなくても  NATトラバーサル動作が必要な対向機器と接続する場合に使用する。なお、通常は
'off'にしておくことが望ましい。
[ノート ]
ipsec ike esp-encapsulation  コマンドとの併用はできない。
また、 IPComp が設定されているトンネルインタフェースでは利用できない。
IKEv1 では、メインモードおよび、アグレッシブモードの  ESP トンネルでのみ利用できる。 AH では利用できず、ト
ランスポートモードでも利用できない。
ただし、 L2TP/IPsec と L2TPv3 を用いた  L2VPN で使用される  IKEv1 では、メインモードかつトランスポートモード
の ESP トンネルでも利用できる。
IKEv2 では、 ESP トンネルを確立する場合のみ利用できる。  AH では利用できず、トランスポートモードでも利用で
きない。330 | コマンドリファレンス  | IPsec の設定

IKEv1メインモードでの NATトラバーサルは、 RTX5000 / RTX3500 Rev.14.00.21 以前、および  RTX1210 Rev.14.01.19
以前では使用不可能。
type オプションは、 RTX5000 / RTX3500 Rev.14.00.21 以前、および  RTX1210 Rev.14.01.19 以前では使用不可能。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•gateway_id
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.48.6 SA の削除
[書式 ]
ipsec  sa delete  id
[設定値及び初期値 ]
•id
• [設定値 ] :
設定値 説明
番号 SA の ID
all すべての  SA
• [初期値 ] : -
[説明 ]
指定した  SA を削除する。
SA の ID は自動的に付与され、 show ipsec sa  コマンドで確認することができる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.49 トンネルインタフェース関連の設定
15.49.1 IPsec トンネルの外側の  IPv4 パケットに対するフラグメントの設定
[書式 ]
ipsec  tunnel  fastpath-fragment-function  follow  df-bit  switch
no ipsec  tunnel  fastpath-fragment-function  follow  df-bit  [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on ESP パケットをフラグメントする必要がある場合に
ESP パケットの  DF ビットに従ってフラグメントす
るかを決定するコマンドリファレンス  | IPsec の設定  | 331

設定値 説明
off ESP パケットをフラグメントする必要がある場合に
ESP パケットの  DF ビットに関係なくフラグメント
する
• [初期値 ] : off
[説明 ]
ESP パケットをフラグメントする必要がある場合に、 DF ビットに従ってフラグメントするか否かを設定する。 ipsec
tunnel outer df-bit コマンドによって  DF ビットがセットされた  ESP パケットであっても本コマンドで  off が設定され
ている場合はフラグメントされる。本コマンドは、トンネルインタフェースに対して設定し、ファストパスで処理
される  ESP パケットのみを対象とする。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.49.2 IPsec トンネルの外側の  IPv4 パケットに対する  DF ビットの制御の設定
[書式 ]
ipsec  tunnel  outer  df-bit  mode
no ipsec  tunnel  outer  df-bit  [mode ]
[設定値及び初期値 ]
•mode
• [設定値 ] :
設定値 説明
copy 内側の  IPv4 パケットの  DF ビットを外側にもコピー
する
set 常に  1
clear 常に  0
• [初期値 ] : copy
[説明 ]
IPsec トンネルの外側の  IPv4 パケットで、 DF ビットをどのように設定するかを制御する。
copy の場合には、内側の  IPv4 パケットの  DF ビットをそのまま外側にもコピーする。
set または  clear の場合には、内側の  IPv4 パケットの  DF ビットに関わらず、外側の  IPv4 パケットの  DF ビットはそ
れぞれ  1、または  0 に設定される。
トンネルインタフェース毎のコマンドである。
[ノート ]
トンネルインタフェースの  MTU と実インタフェースの  MTU の値の大小関係により、 IPsec 化されたパケットをフ
ラグメントしなくてはいけない時には、このコマンドの設定に関わらず  DF ビットは  0 になる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.49.3 使用する  SA のポリシーの設定
[書式 ]
ipsec  tunnel  policy_id
no ipsec  tunnel  [policy_id ]
[設定値及び初期値 ]
•policy_id
• [設定値 ] : 整数  (1..2147483647)
• [初期値 ] : -332 | コマンドリファレンス  | IPsec の設定

[説明 ]
選択されているトンネルインタフェースで使用する  SA のポリシーを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.49.4 IPComp によるデータ圧縮の設定
[書式 ]
ipsec  ipcomp  type type
no ipsec  ipcomp  type [type]
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
deflate deflate 圧縮でデータを圧縮する
none データ圧縮を行わない
• [初期値 ] : none
[説明 ]
IPComp でデータ圧縮を行うかどうかを設定する。サポートしているアルゴリズムは  deflate のみである。
受信した  IPComp パケットを展開するためには、特別な設定を必要としない。すなわち、サポートしているアルゴリ
ズムで圧縮された  IPComp パケットを受信した場合には、設定に関係なく展開する。
必ずしもセキュリティ・ゲートウェイの両方にこのコマンドを設定する必要はない。片側にのみ設定した場合には、
そのセキュリティ・ゲートウェイから送信される  IP パケットのみが圧縮される。
トランスポートモードのみを使用する場合には、 IPComp を使用することはできない。
[ノート ]
データ圧縮には、 PPP で使われる  CCP や、フレームリレーで使われる  FRF.9 もある。圧縮アルゴリズムとして、
IPComp で使われる  deflate と、 CCP/FRF.9 で使われる  Stac-LZS との間に基本的な違いはない。しかし、 CCP/FRF.9
でのデータ圧縮は  IPsec による暗号化の後に行われる。このため、 暗号化でランダムになったデータを圧縮しようと
することになり、ほとんど効果がない。一方、 IPComp は IPsec による暗号化の前にデータ圧縮が行われるため、一
定の効果が得られる。また、 CCP/FRF.9 とは異なり、対向のセキュリティ．ゲートウェイまでの全経路で圧縮され
たままのデータが流れるため、例えば本機の出力インタフェースが  LAN であってもデータ圧縮効果を期待できる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.49.5 トンネルバックアップの設定
[書式 ]
tunnel  backup  none
tunnel  backup  interface  ip_address
tunnel  backup  pp peer_num  [switch-router= switch1 ]
tunnel  backup  tunnel  tunnel_num  [switch-interface= switch2 ]
no tunnel  backup
[設定値及び初期値 ]
• none : トンネルバックアップを使用しない
• [初期値 ] : none
•interface
• [設定値 ] : LAN インタフェース名
• [初期値 ] : -
•ip_address
• [設定値 ] : バックアップ先のゲートウェイの  IP アドレス
• [初期値 ] : -
•peer_numコマンドリファレンス  | IPsec の設定  | 333

• [設定値 ] : バックアップ先の相手先情報番号
• [初期値 ] : -
•tunnel_num
• [設定値 ] : トンネルインタフェース番号
• [初期値 ] : -
•switch1  : バックアップの受け側のルーターを  2 台に分けるか否か
• [設定値 ] :
設定値 説明
on 分ける
off 分けない
• [初期値 ] : off
•switch2  : LAN/PP インタフェースのバックアップにしたがってトンネルを作り直すか否か
• [設定値 ] :
設定値 説明
on 作り直す
off 作り直さない
• [初期値 ] : on
[説明 ]
トンネルインタフェースに障害が発生したときにバックアップとして利用するインタフェースを指定する。
switch-router オプションについては、以下の  2 つの条件を満たすときに  on を設定する。
•バックアップの受け側に  2 台のルーターがあり、一方がバックアップ元の回線に接続し、もう一方がバックアッ
プ先の回線に接続している。
•バックアップ先の回線に接続している  ヤマハルーター  のファームウェアまたはイメージが、このリビジョンよ
りも古い。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•tunnel_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
•peer_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 150
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.49.6 トンネルテンプレートの設定
[書式 ]
tunnel  template  tunnel_num  [tunnel_num  ...]
no tunnel  template334 | コマンドリファレンス  | IPsec の設定

[設定値及び初期値 ]
•tunnel_num
• [設定値 ] : トンネルインタフェース番号、 または間にハイフン  (-) をはさんでトンネルインタフェース番号を範
囲指定したもの
• [初期値 ] : -
[説明 ]
tunnel select コマンドにて選択されたトンネルインタフェースを展開元として、当該インタフェースに設定されてい
るコマンドの展開先となるトンネルインタフェースを設定する。
展開元のトンネルインタフェースに設定することで、展開先のトンネルインタフェースにも適用されるコマンドは
以下のとおりである。なお、末尾に  (*1) が付加されているコマンドについては  [ノート ] を参照のこと。
•ipsec tunnel
•ipsec sa policy
•ipsec ike で始まるコマンドのうち、パラメータにセキュリティ・ゲートウェイの識別子をとるもの
•ipsec auto refresh  (引数にセキュリティ・ゲートウェイの識別子を指定する場合 )
•tunnel encapsulation
•tunnel ngn arrive permit
•tunnel ngn bandwidth
•tunnel ngn disconnect time
•tunnel ngn radius auth
•l2tpで始まるコマンド  (*1)
•tunnel enable
上記コマンドのうち以下のコマンドについては、特定のパラメータの値が展開元のトンネルインタフェース番号に
一致する場合のみ、コマンドが展開される。その場合、当該パラメータの値は展開先のトンネルインタフェース番
号に置換される。
コマンド パラメータ
ipsec tunnel ポリシー ID
ipsec sa policy ポリシー ID
ipsec ike で始まるコマンド セキュリティ・ゲートウェイの識別子
ipsec auto refresh セキュリティ・ゲートウェイの識別子
tunnel enable トンネルインタフェース番号
ipsec sa policy コマンドでは、セキュリティ・ゲートウェイの識別子が展開先のトンネルインタフェース番号に置換
される。
ipsec ike remote name コマンドでは、相手側セキュリティ・ゲートウェイの名前の末尾に展開先のトンネルインタフ
ェース番号が付加される。
展開元のトンネルインタフェースに設定されているコマンドと同じコマンドが、展開先のトンネルインタフェース
に既に設定されている場合、展開先のトンネルインタフェースに設定されているコマンドが優先される。
コマンド展開後の、ルーターの動作時に参照される設定は show config tunnel コマンドに expandキーワードを指定
することで確認できる。
[ノート ]
トンネルインタフェースが選択されている時にのみ使用できる。
展開対象となるコマンドのうち、末尾に  (*1) が付加されているコマンドは、 RTX5000 / RTX3500 Rev.14.00.11 以前で
は対応していない。
[設定例 ]
展開先のトンネルインタフェースとして、番号の指定と範囲の指定を同時に記述することができる。
tunnel select 1
 tunnel template 8 10-20
tunnel select 2
 tunnel template 100 200-300 400
以下の 2つの設定は同じ内容を示している。
tunnel select 1コマンドリファレンス  | IPsec の設定  | 335

 tunnel template 2
  ipsec tunnel 1
  ipsec sa policy 1 1 esp aes-cbc sha-hmac
  ipsec ike encryption 1 aes-cbc
  ipsec ike group 1 modp1024
  ipsec ike local address 1 192.168.0.1
  ipsec ike pre-shared-key 1 text himitsu1
  ipsec ike remote address 1 any
  ipsec ike remote name 1 pc
 tunnel enable 1
tunnel select 2
  ipsec ike pre-shared-key 2 text himitsu2
tunnel select 1
 ipsec tunnel 1
  ipsec sa policy 1 1 esp aes-cbc sha-hmac
  ipsec ike encryption 1 aes-cbc
  ipsec ike group 1 modp1024
  ipsec ike local address 1 192.168.0.1
  ipsec ike pre-shared-key 1 text himitsu1
  ipsec ike remote address 1 any
  ipsec ike remote name 1 pc
 tunnel enable 1
tunnel select 2
 ipsec tunnel 2
  ipsec sa policy 2 2 esp aes-cbc sha-hmac
  ipsec ike encryption 2 aes-cbc
  ipsec ike group 2 modp1024
  ipsec ike local address 2 192.168.0.1
  ipsec ike pre-shared-key 2 text himitsu2
  ipsec ike remote address 2 any
  ipsec ike remote name 2 pc2
 tunnel enable 2
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•tunnel_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.50 トランスポートモード関連の設定
15.50.1 トランスポートモードの定義
[書式 ]
ipsec  transport  id policy_id  [proto  [src_port_list  [dst_port_list ]]]
no ipsec  transport  id [policy_id  [proto  [src_port_list  [dst_port_list ]]]]
[設定値及び初期値 ]
•id
• [設定値 ] : トランスポート  ID(1..2147483647)
• [初期値 ] : -
•policy_id
• [設定値 ] : ポリシー ID(1..2147483647)
• [初期値 ] : -
•proto336 | コマンドリファレンス  | IPsec の設定

• [設定値 ] : プロトコル
• [初期値 ] : -
•src_port_list  : UDP、TCP のソースポート番号列
• [設定値 ] :
•ポート番号を表す十進数
•ポート番号を表すニーモニック
• *( すべてのポート  )
• [初期値 ] : -
•dst_port_list  : UDP、TCP のデスティネーションポート番号列
• [設定値 ] :
•ポート番号を表す十進数
•ポート番号を表すニーモニック
• *( すべてのポート  )
• [初期値 ] : -
[説明 ]
トランスポートモードを定義する。
定義後、 proto、src_port_list 、dst_port_list  パラメータに合致する  IP パケットに対してトランスポートモードでの通
信を開始する。
[ノート ]
TCP プロトコルは  vRX シリーズ  では指定不可能。
[設定例 ]
• TELNET のデータをトランスポートモードで通信
# ipsec sa policy 101 1 esp aes-cbc sha-hmac
# ipsec transport 1 101 tcp * telnet
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.50.2 トランスポートモードのテンプレートの設定
[書式 ]
ipsec  transport  template  id1 id2 [id2 ...]
no ipsec  transport  template  id1 [id2 ...]
[設定値及び初期値 ]
•id1
• [設定値 ] : 展開元のトランスポート  ID
• [初期値 ] : -
•id2
• [設定値 ] : 展開先のトランスポート  ID 、または間にハイフン (-)をはさんでトランスポート  ID を範囲指定した
もの
• [初期値 ] : -
[説明 ]
指定した  ipsec transport  コマンドの設定の展開先となるトランスポート  ID を設定する。展開先のポリシー  ID は展
開先のトランスポート  ID と同じ値が設定される。
展開先のトランスポート  ID に対して既に設定が存在する場合、展開先の設定が優先される。
本コマンドによって VPN対地数まで  ipsec transport  コマンドの設定を展開することができる。
VPN対地数を超える範囲に展開することはできない。コマンドリファレンス  | IPsec の設定  | 337

[ノート ]
RTX5000 は Rev.14.00.12 以降で使用可能。
RTX3500 は Rev.14.00.12 以降で使用可能。
[設定例 ]
展開先の設定としてトランスポート  ID とトランスポート  ID の範囲を同時に記述することができる。
ipsec transport 1 1 udp 1701 *
ipsec transport template 1 10 20-30
以下の 2つの設定は同じ内容を示している。
ipsec transport 1 1 udp 1701 *
ipsec transport template 1 2 10-12
ipsec transport 1 1 udp 1701 *
ipsec transport 2 2 udp 1701 *
ipsec transport 10 10 udp 1701 *
ipsec transport 11 11 udp 1701 *
ipsec transport 12 12 udp 1701 *
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、 ipsec transport  コマンドの設定を展開するトランスポート  ID ( id) の最大個数
が拡張される。
•id
ライセンス名拡張後の最大個数
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.51 PKI 関連の設定
15.51.1 証明書ファイルの設定
[書式 ]
pki certificate  file cert_id  file type [password ]
no pki certificate  file cert_id  [file ...]
[設定値及び初期値 ]
•cert_id
• [設定値 ] :
設定値 説明
1..8 証明書ファイルの識別子
• [初期値 ] : -
•file
• [設定値 ] : 証明書ファイルの絶対パスまたは相対パス
• [初期値 ] : -
•type : ファイル形式338 | コマンドリファレンス  | IPsec の設定

• [設定値 ] :
設定値 説明
pkcs12 PKCS#12 形式のファイル
x509-pem X.509 PEM 形式のファイル
• [初期値 ] : -
•password
• [設定値 ] : ファイルを復号するためのパスワード (半角 64文字以内 )
• [初期値 ] : -
[説明 ]
証明書ファイルを設定する。
file に相対パスを指定する場合、 set コマンドの環境変数  pwd で指定したディレクトリからの相対パスを指定する。
type に pkcs12 を指定した場合、ファイルを復号するための  password  を指定する必要がある。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
15.51.2 CRL ファイルの設定
[書式 ]
pki crl file crl_id  file
no pki crl file crl_id  [file]
[設定値及び初期値 ]
•crl_id
• [設定値 ] :
設定値 説明
1..8 CRLファイルの識別子
• [初期値 ] : -
•file
• [設定値 ] : CRLファイルの絶対パスまたは相対パス
• [初期値 ] : -
[説明 ]
CRLファイルを設定する。
file に相対パスを指定する場合、 set コマンドの環境変数  pwd で指定したディレクトリからの相対パスを指定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | IPsec の設定  | 339

