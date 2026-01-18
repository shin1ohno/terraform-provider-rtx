# 第6章: ISDN 関連の設定

> 元PDFページ: 122-136

---

第 6 章
ISDN 関連の設定
6.1 共通の設定
6.1.1 BRI 回線の種類の指定
[書式 ]
line type interface  line [channels ]
no line type interface  line [channels ]
[設定値及び初期値 ]
•interface
• [設定値 ] : BRI インタフェース名
• [初期値 ] : -
•line
• [設定値 ] :
設定値 説明
isdn, isdn-ntt ISDN 回線交換
l64 ディジタル専用線、 64kbit/s
l128 ディジタル専用線、 128kbit/s
• [初期値 ] : isdn
•channels  : line パラメータが  isdn､ isdn-ntt の場合のみ指定可
• [設定値 ] :
設定値 説明
1b B チャネルは  1 チャネルだけ使用
2b B チャネルは  2 チャネルとも使用する
• [初期値 ] : 2b
[説明 ]
BRI 回線の種類を指定する。設定の変更は、再起動か、あるいは該当インタフェースに対する interface reset  コマン
ドの発行により反映される。
[ノート ]
別の通信機器の発着信のために  1B チャネルを確保したい場合は channels  パラメータを  1b に設定する。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.1.2 自分の  ISDN 番号の設定
[書式 ]
isdn local  address  interface  isdn_num [/sub_address ]
isdn local  address  interface  [/sub_address ]
no isdn local  address  interface
[設定値及び初期値 ]
•interface
• [設定値 ] :
• BRI インタフェース名
• PRI インタフェース名
• [初期値 ] : -
•isdn_num122 | コマンドリファレンス  | ISDN 関連の設定

• [設定値 ] : ISDN 番号
• [初期値 ] : -
•sub_address
• [設定値 ] : ISDN サブアドレス  (0x21 から  0x7e の ASCII 文字列  )
• [初期値 ] : -
[説明 ]
自分の  ISDN 番号とサブアドレスを設定する。 ISDN 番号、サブアドレスとも完全に設定して運用することが推奨さ
れる。また、 ISDN 番号は市外局番も含めて設定する。
[ノート ]
他機種との相互接続のために、 ISDN サブアドレスに英文字や記号を使わず数字だけにしなければいけないことがあ
る
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.1.3 終端抵抗の設定
[書式 ]
isdn terminator  interface  terminator
no isdn terminator  interface  [terminator ]
[設定値及び初期値 ]
•interface
• [設定値 ] : BRI インタフェース名
• [初期値 ] : -
•terminator
• [設定値 ] :
設定値 説明
on 終端抵抗を  ON にする
off 終端抵抗を  OFF にする
• [初期値 ] : on
[説明 ]
指定した  BRI インタフェースの終端抵抗を  ON または  OFF にする。
[ノート ]
ルーターを外部 DSUに直結する場合にはルーターの終端抵抗を必ず ONにする。 DSUからのバス配線で接続する
場合には、通常はバス配線に終端抵抗が配置されているので、ルーターの終端抵抗は OFFにする。ただし、ルータ
ーがバス配線の終端にあり、バス配線内に終端抵抗が配置されていないときには、ルータの終端抵抗を ONにしな
ければならない。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.1.4 PP で使用するインタフェースの設定
[書式 ]
pp bind  interface  [interface ]
no pp bind  [interface ]
[設定値及び初期値 ]
•interface
• [設定値 ] : BRI インタフェース名と  BRI インタフェース名の並び
• [初期値 ] : -
[説明 ]
選択されている相手先に対して実際に使用するインタフェースを設定する。コマンドリファレンス  | ISDN 関連の設定  | 123

[適用モデル ]
RTX5000, RTX3500, RTX1210
6.1.5 課金額による発信制限の設定
[書式 ]
account  threshold  [interface ] yen
account  threshold  pp yen
no account  threshold  interface  [yen]
no account  threshold  [yen]
no account  threshold  pp [yen]
[設定値及び初期値 ]
•interface
• [設定値 ] :
• BRI インタフェース名
• PRI インタフェース名
• [初期値 ] : -
•yen
• [設定値 ] :
設定値 説明
円 (1..2147483647) 課金額
off 発信制限機能を使わない
• [初期値 ] : off
[説明 ]
網から通知される課金の合計  ( これは show account  コマンドで表示される  ) の累計が指定した金額に達したらそれ
以上の発信を行わないようにする。
account threshold  コマンドではルーター全体の合計金額を設定し、 interface  パラメータを指定した場合には、それぞ
れのインタフェースでの合計金額、 account threshold pp  コマンドでは選択している相手先に対する発信での合計金
額で制御を行う。
課金が網から通知されるのは通信切断時なので、長時間の接続の途中切断することはできず、この場合は制限はで
きない。この場合に対処するには、 isdn forced disconnect time  コマンドで通信中でも時間を監視して強制的に回線
を切るような設定にしておく方法がある。また、課金合計は clear account  コマンドで  0 にリセットでき、 schedule
at コマンドで定期的に clear account  を実行するようにしておくと、毎月一定額以内に課金を抑えるといったことが
自動で可能になる。
[ノート ]
電源  OFF や再起動により、それまでの課金情報がクリアされることに注意。課金額は通信の切断時に  NTT から
ISDN で通知される料金情報に基づくため、割引サービスなどを利用している場合には、最終的に  NTT から請求さ
れる料金とは異なる場合がある。また、 NTT 以外の通信事業者を利用して通信した場合には料金情報は通知されな
い。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.1.6 専用線がダウンした時にバックアップする相手先情報番号の設定
[書式 ]
leased  backup  peer_num
no leased  backup  [peer_num ]
[設定値及び初期値 ]
•peer_num
• [設定値 ] :124 | コマンドリファレンス  | ISDN 関連の設定

設定値 説明
番号 バックアップする相手先情報番号
none ISDN でバックアップをしない
• [初期値 ] : none
[説明 ]
BRI インタフェースを複数持つ機種で有効なコマンド。
選択した相手先に対する専用線がダウンした場合に  ISDN でバックアップする、 バックアップ用の相手先情報番号を
設定する。
[適用モデル ]
RTX5000, RTX3500
6.2 相手側の設定
6.2.1 常時接続の設定
[書式 ]
pp always-on  switch  [time]
no pp always-on
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on 常時接続する
off 常時接続しない
• [初期値 ] : off
•time
• [設定値 ] : 再接続を要求するまでの秒数  (60..21474836)
• [初期値 ] : -
[説明 ]
選択されている相手について常時接続するか否かを設定する。また、常時接続での通信終了時に再接続を要求する
までの時間間隔を指定する。
常時接続に設定されている場合には、起動時に接続を起動し、通信終了時には再接続を起動し、キープアライブ機
能により接続相手のダウン検出を行う。接続失敗時あるいは通信の異常終了時には time に設定された時間間隔を待
った後に再接続の要求を行い、正常な通信終了時には直ちに再接続の要求を行う。 switch  が on に設定されている場
合には、 time の設定が有効となる。 time が設定されていない場合、 time は 60 になる。
以下のコマンドが設定されている場合、 switch  を on に設定した時点で接続処理が行われる。
• PPPoE 接続
•pppoe use
•pp enable
• ISDN 接続
•pp bind  BRI インタフェース名
•pp enable
•モバイルインターネット接続  ( 携帯端末を  PP (USB モデム ) として制御するタイプ  )
•pp bind  usb1
•pp enable
•mobile use
また、上記の設定に依らず、 switch  を off に設定した時点で切断処理が行われる。コマンドリファレンス  | ISDN 関連の設定  | 125

[ノート ]
PP 毎のコマンドである。
PP として専用線が使用される時、あるいは  anonymous が選択された時には無効である。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
6.2.2 相手  ISDN 番号の設定
[書式 ]
isdn remote  address  call_arrive  isdn_num  [/sub_address ] [isdn_num_list ]
isdn remote  address  call_arrive  isdn_num  [isdn_num_list ]
no isdn remote  address  call_arrive  [isdn_num  [/sub_address ] [isdn_num_list ]]
[設定値及び初期値 ]
•call_arrive
• [設定値 ] :
設定値 説明
call 発着信用
arrive 着信専用
• [初期値 ] : -
•isdn_num
• [設定値 ] : ISDN 番号
• [初期値 ] : -
•sub_address
• [設定値 ] : ISDN サブアドレス  (0x21 から  0x7e の ASCII 文字  )
• [初期値 ] : -
•isdn_num_list
• [設定値 ] : ISDN 番号だけまたは  ISDN 番号とサブアドレスの組を空白で区切った並び
• [初期値 ] : -
[説明 ]
選択されている相手の  ISDN 番号とサブアドレスを設定する。 ISDN 番号には市外局番も含めて設定する。
選択されている相手が  anonymous の場合は無意味である。
複数の  ISDN 番号が設定されている場合、まず先頭の  ISDN 番号での接続に失敗すると次に指定された  ISDN 番号が
使われる。同様に、それに失敗すると次の  ISDN 番号を使うという動作を続ける。
MP のように相手先に対して複数チャネルで接続しようとする際に発信する順番は、 isdn remote call order  コマンド
で設定する。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.3 自動接続の設定
[書式 ]
isdn auto  connect  auto
no isdn auto  connect  [auto]
[設定値及び初期値 ]
•auto
• [設定値 ] :
設定値 説明
on 自動接続する
off 自動接続しない126 | コマンドリファレンス  | ISDN 関連の設定

• [初期値 ] : on
[説明 ]
選択されている相手について自動接続するか否かを設定する。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.4 相手への発信順序の設定
[書式 ]
isdn remote  call order  order
no isdn remote  call order  [order ]
[設定値及び初期値 ]
•order
• [設定値 ] :
設定値 説明
round ラウンドロビン方式
serial 順次サーチ方式
• [初期値 ] : serial
[説明 ]
isdn remote address call  コマンドで複数の  ISDN 番号が設定されている場合に意味を持つ。 MP を使用する場合など
のように、相手先に対して同時に複数のチャネルで接続しようとする際に、どのような順番で  ISDN 番号を選択する
かを設定する。
round を指定した場合は、 isdn remote address call  コマンドで最初に設定した  ISDN 番号で発信した次の発信時に、
このコマンドで次に設定された  ISDN 番号を使う。このように順次ずれていき、 最後に設定された番号で発信した次
には、最初に設定された  ISDN 番号を使い、これを繰り返す。
serial を指定した場合は、発信時には必ず最初に設定された  ISDN 番号を使い、何らかの理由で接続できなかった場
合は次に設定された  ISDN 番号で発信し直す。
なお  round、 serial いずれの設定の場合でも、どことも接続されていない状態や相手先とすべてのチャネルで切断さ
れた後では、最初に設定された  ISDN 番号から発信に使用される。
[ノート ]
MP を使用する場合は、 round にした方が効率がよい。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.5 着信許可の設定
[書式 ]
isdn arrive  permit  arrive  [vrrp interface  vrid[slave]]
no isdn arrive  permit  [arrive ]
[設定値及び初期値 ]
•arrive
• [設定値 ] :
設定値 説明
on 許可する
off 許可しない
• [初期値 ] : on
•interface
• [設定値 ] : LAN インタフェース名
• [初期値 ] : -
•vridコマンドリファレンス  | ISDN 関連の設定  | 127

• [設定値 ] : VRRP グループ  ID(1..255)
• [初期値 ] : -
[説明 ]
選択されている相手からの着信を許可するか否かを設定する。
on に設定しかつ  VRRP グループを指定することで、 VRRP の状態によって着信を許可するか否かの動作を動的に変
えることが可能である。
この時、 slave パラメータを省略した場合には指定した  VRRP グループでマスターとして動作している場合にのみ着
信が許可される。 slave パラメータを設定した場合には、指定した  VRRP グループで非マスターである場合にのみ着
信が許可される。
[ノート ]
isdn arrive permit 、isdn call permit  コマンドとも  off を設定した場合、 ISDN 回線経由では通信できない。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.6 発信許可の設定
[書式 ]
isdn call permit  permit
no isdn call permit  [permit ]
[設定値及び初期値 ]
•permit
• [設定値 ] :
設定値 説明
on 許可する
off 許可しない
• [初期値 ] : on
[説明 ]
選択されている相手への発信を許可するか否かを設定する。
[ノート ]
isdn arrive permit 、isdn call permit  コマンドとも  off を設定した場合、 ISDN 回線経由では通信できない。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.7 再発信抑制タイマの設定
[書式 ]
isdn call block  time  time
no isdn call block  time  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : 秒数  (0..15.0)
• [初期値 ] : 0
[説明 ]
選択されている相手との通信が切断された後、同じ相手に対し再度発信するのを禁止する時間を設定する。秒数は
0.1 秒単位で設定できる。
isdn call prohibit time  コマンドによるタイマはエラーで切断された場合だけに適用されるが、 このコマンドによるタ
イマは正常切断でも適用される点が異なる。128 | コマンドリファレンス  | ISDN 関連の設定

[ノート ]
切断後すぐに発信ということを繰り返す状況では適当な値を設定すべきである。
isdn forced disconnect time  コマンドと併用するとよい。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.8 エラー切断後の再発信禁止タイマの設定
[書式 ]
isdn call prohibit  time  time
no isdn call prohibit  time  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : 秒数  (60..21474836.0)
• [初期値 ] : 60
[説明 ]
選択されている相手に発信しようとして失敗した場合に、同じ相手に対し再度発信するのを禁止する時間を設定す
る。秒数は  0.1 秒単位で設定できる。
isdn call block time  コマンドによるタイマは切断後に常に適用されるが、このコマンドによるタイマはエラー切断に
のみ適用される点が異なる。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.9 相手にコールバック要求を行うか否かの設定
[書式 ]
isdn callback  request  callback_request
no isdn callback  request  [callback_request ]
[設定値及び初期値 ]
•callback_request
• [設定値 ] :
設定値 説明
on 要求する
off 要求しない
• [初期値 ] : off
[説明 ]
選択されている相手に対してコールバック要求を行うか否かを設定する。
[ノート ]
コールバックは、 backupコマンドによる ISDNインタフェースへのバックアップとの併用はできません。  バックア
ップと併用する場合は、ネットワークバックアップを使ってください。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.10 相手からのコールバック要求に応じるか否かの設定
[書式 ]
isdn callback  permit  callback_permit
no isdn callback  permit  [callback_permit ]
[設定値及び初期値 ]
•callback_permitコマンドリファレンス  | ISDN 関連の設定  | 129

• [設定値 ] :
設定値 説明
on 応じる
off 応じない
• [初期値 ] : off
[説明 ]
選択されている相手からのコールバック要求に対してコールバックするか否かを設定する。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.11 コールバック要求タイプの設定
[書式 ]
isdn callback  request  type type
no isdn callback  request  type [type]
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
yamaha ヤマハ方式
mscbcp MS コールバック
• [初期値 ] : yamaha
[説明 ]
コールバックを要求する場合のコールバック方式を設定する。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.12 コールバック受け入れタイプの設定
[書式 ]
isdn callback  permit  type type1  [type2 ]
no isdn callback  permit  type [type1  [type2 ]]
[設定値及び初期値 ]
•type1,type2
• [設定値 ] :
設定値 説明
yamaha ヤマハ方式
mscbcp MS コールバック
• [初期値 ] :
• type1=yamaha
• type2=mscbcp
[説明 ]
受け入れることのできるコールバック方式を設定する。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.13 MS コールバックでユーザからの番号指定を許可するか否かの設定
[書式 ]
isdn callback  mscbcp  user-specify  specify130 | コマンドリファレンス  | ISDN 関連の設定

no no isdn callback  mscbcp  user-specify  [specify ]
[設定値及び初期値 ]
•specify
• [設定値 ] :
設定値 説明
on 許可する
off 拒否する
• [初期値 ] : off
[説明 ]
サーバー側として動作する場合にはコールバックするために利用可能な電話番号が一つでもあればそれに対しての
みコールバックする。しかし、 anonymous への着信で、発信者番号通知がなく、コールバックのためにつかえる電
話番号が全く存在しない場合に、 コールバック要求側  ( ユーザ  ) からの番号指定によりコールバックするかどうかを
設定する。
[ノート ]
設定が  off でコールバックできない場合には、コールバックせずにそのまま接続する。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.14 コールバックタイマの設定
[書式 ]
isdn callback  response  time  type time
no isdn callback  response  time  [type]
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
1b 1B でコールバックする
• [初期値 ] : -
•time
• [設定値 ] : 秒数  (0..15.0)
• [初期値 ] : 0
[説明 ]
選択されている相手からのコールバック要求を受け付けてから、実際に相手に発信するまでの時間を設定する。
秒数は  0.1 秒単位で設定できる。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.15 コールバック待機タイマの設定
[書式 ]
isdn callback  wait time  time
no isdn callback  wait time  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : 秒数  (1..60.0)
• [初期値 ] : 60コマンドリファレンス  | ISDN 関連の設定  | 131

[説明 ]
選択されている相手にコールバックを要求し、それが受け入れられていったん回線が切断されてから、このタイマ
がタイムアウトするまで相手からのコールバックによる着信を受け取れなかった場合には接続失敗とする。秒数は
0.1 秒単位で設定できる。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.16 ISDN 回線を切断するタイマ方式の指定
[書式 ]
isdn disconnect  policy  type
no isdn disconnect  policy  [type]
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
1 単純トラフィック監視方式
2 課金単位時間方式
• [初期値 ] : 1
[説明 ]
単純トラフィック監視方式は従来型の方式であり、 isdn disconnect time 、isdn disconnect input time 、isdn disconnect
output time  の 3 つのタイマコマンドでトラフィックを監視し、一定時間パケットが流れなくなった時点で回線を切
断する。
課金単位時間方式では、課金単位時間と監視時間を isdn disconnect interval time  コマンドで設定し、監視時間中にパ
ケットが流れなければ課金単位時間の倍数の時間で回線を切断する。通信料金を減らす効果が期待できる。
[設定例 ]
# isdn disconnect policy 2
# isdn disconnect interval time 240 6 2
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.17 切断タイマの設定  ( ノーマル  )
[書式 ]
isdn disconnect  time  time
no isdn disconnect  time  [time]
[設定値及び初期値 ]
•time
• [設定値 ] :
設定値 説明
1..21474836.0 秒数
off タイマを設定しない
• [初期値 ] : 60
[説明 ]
選択されている相手について  PP 側のデータ送受信がない場合の切断までの時間を設定する。秒数は  0.1 秒単位で
設定できる。
[ノート ]
本コマンドを offに設定した場合には、 isdn disconnect input time  コマンドおよび isdn disconnect output time  コマン
ドの設定にかかわらず切断されなくなる。132 | コマンドリファレンス  | ISDN 関連の設定

本コマンドの設定値を  NORMAL 秒、 isdn disconnect input time  コマンドの設定値を  IN 秒、 isdn disconnect output
time  コマンドの設定値を  OUT 秒とする。
NORMAL>IN または  OUT>NORMAL のように設定した場合、設定値が大きい方が優先される。そのため、パケット
の入力が観測されないと  NORMAL 秒、パケットの出力が観測されないと  OUT 秒で切断される。なお、パケットの
入出力が観測されないと常に  NORMAL 秒で切断される。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.18 切断タイマの設定  ( ファスト  )
[書式 ]
isdn fast disconnect  time  time
no no isdn fast disconnect  time  [time]
[設定値及び初期値 ]
•time
• [設定値 ] :
設定値 説明
1..21474836.0 秒数
off タイマを設定しない
• [初期値 ] : 20
[説明 ]
ある宛先について、パケットがルーティングされ、そこへ発信しようとしたが、 ISDN 回線が他の接続先により塞が
っていて発信できない場合に、 ISDN 回線を塞いでいる相手先についてこのタイマが動作を始める。このタイマで指
定した時間の間、パケットが全く流れなかったらその相手先を切断して、発信待ちの宛先を接続する。秒数は  0.1 秒
単位で設定できる。
なお、 isdn auto connect  コマンドが  off の場合はこのタイマは無視される。
[ノート ]
同じ  ISDN 回線に接続されている他の機器が  Bch を使用している場合には、本コマンドは機能しないことがある。
また、本機の  PP Anonymous の接続がすべての  Bch を使用している場合には、新たな  PP Anonymous の接続を起動し
ても、本コマンドは機能しない。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.19 切断タイマの設定  ( 強制  )
[書式 ]
isdn forced  disconnect  time  time
no isdn forced  disconnect  time  [time]
[設定値及び初期値 ]
•time
• [設定値 ] :
設定値 説明
1..21474836.0 秒数
off タイマを設定しない
• [初期値 ] : off
[説明 ]
選択されている相手に接続する最大時間を設定する。秒数は  0.1 秒単位で設定できる。
パケットをやりとりしていても、このコマンドで設定した時間が経過すれば強制的に回線を切断する。コマンドリファレンス  | ISDN 関連の設定  | 133

ダイヤルアップ接続でインターネット側からの無効なパケット  (ping アタック等  ) が原因で回線が自動切断できな
い場合に有効。 isdn call block time  コマンドと併用するとよい。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.20 入力切断タイマの設定  ( ノーマル  )
[書式 ]
isdn disconnect  input  time  time
no isdn disconnect  input  time  [time]
[設定値及び初期値 ]
•time
• [設定値 ] :
設定値 説明
1..21474836.0 秒数
off タイマを設定しない
• [初期値 ] : 120
[説明 ]
選択されている相手について  PP 側からデータ受信がない場合の切断までの時間を設定する。秒数は  0.1 秒単位で
設定できる。
[ノート ]
例えば、 UDP パケットを定期的に出すようなプログラムが暴走したような場合、本タイマを設定しておくことによ
り回線を切断することができる。
isdn disconnect time  コマンドを offに設定した場合には、 本コマンドおよび isdn disconnect output time  コマンドの設
定にかかわらず切断されなくなる。
isdn disconnect time  コマンドの設定値を  NORMAL 秒、本コマンドの設定値を  IN 秒、 isdn disconnect output time  コ
マンドの設定値を  OUT 秒とする。
NORMAL>IN または  OUT>NORMAL のように設定した場合、設定値が大きい方が優先される。そのため、パケット
の入力が観測されないと  NORMAL 秒、パケットの出力が観測されないと  OUT 秒で切断される。なお、パケットの
入出力が観測されないと常に  NORMAL 秒で切断される。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.21 出力切断タイマの設定  ( ノーマル  )
[書式 ]
isdn disconnect  output  time  time
no isdn disconnect  output  time  [time]
[設定値及び初期値 ]
•time
• [設定値 ] :
設定値 説明
1..21474836.0 秒数
off タイマを設定しない
• [初期値 ] : 120
[説明 ]
選択されている相手について  PP 側へのデータ送信がない場合の切断までの時間を設定する。秒数は  0.1 秒単位で
設定できる。134 | コマンドリファレンス  | ISDN 関連の設定

[ノート ]
例えば、接続先を経由して外部から不正な UDP パケットを受信し続けるような場合、本タイマを設定しておくこと
により回線を切断することができる。
isdn disconnect time  コマンドを offに設定した場合には、 isdn disconnect input time  コマンドおよび本コマンドの設
定にかかわらず切断されなくなる。
isdn disconnect time  コマンドの設定値を  NORMAL 秒、isdn disconnect input time  コマンドの設定値を  IN 秒、本コマ
ンドの設定値を  OUT 秒とする。
NORMAL>IN または  OUT>NORMAL のように設定した場合、設定値が大きい方が優先される。そのため、パケット
の入力が観測されないと  NORMAL 秒、パケットの出力が観測されないと  OUT 秒で切断される。なお、パケットの
入出力が観測されないと常に  NORMAL 秒で切断される。
[適用モデル ]
RTX5000, RTX3500, RTX1210
6.2.22 課金単位時間方式での課金単位時間と監視時間の設定
[書式 ]
isdn disconnect  interval  time  unit watch  spare
no isdn disconnect  interval  time  [unit watch  spare ]
[設定値及び初期値 ]
•unit : 課金単位時間
• [設定値 ] :
•秒数  (1..21474836.0)
• off
• [初期値 ] : 180
•watch  : 監視時間
• [設定値 ] :
•秒数  (1..21474836.0)
• off
• [初期値 ] : 6
•spare  : 切断余裕時間
• [設定値 ] :
•秒数  (1..21474836.0)
• off
• [初期値 ] : 2
[説明 ]
課金単位時間方式で使われる、課金単位時間と監視時間を設定する。秒数は  0.1 秒単位で設定できる。
それぞれの意味は下図参照。
watch  で示した間だけトラフィックを監視し、この間にパケットが流れなければ回線を切断する。 spare  は切断処理
に時間がかかりすぎて、実際の切断が単位時間を越えないように余裕を持たせるために使う。
回線を接続している時間が unit の倍数になるので、単純トラフィック監視方式よりも通信料金を減らす効果が期待
できる。
[設定例 ]
# isdn disconnect policy 2
# isdn disconnect interval time 240 6 2コマンドリファレンス  | ISDN 関連の設定  | 135

[適用モデル ]
RTX5000, RTX3500, RTX1210136 | コマンドリファレンス  | ISDN 関連の設定

