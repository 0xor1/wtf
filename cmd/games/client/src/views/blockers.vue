<template>
  <div class="root">
    <div v-if="loading">
      loading...
    </div>
    <div v-if="errors.length > 0">
      <p v-for="(error, index) in errors" :key="index">{{error}}</p>
    </div>
    <div v-if="game != null && game.board != null" class="game">
      <table class="board" @touchstart="touchBoard" @touchmove="touchBoard" @click.stop.prevent="place" @contextmenu.stop.prevent="rotate" @wheel="flip">
        <tr v-for="(_, y) in boardDims" :key="y">
          <td v-for="(_, x) in boardDims" :key="x" class="cell" :data-x="x" :data-y="y" :class="boardCellClass(x, y)" @mouseenter.stop.prevent="move(x, y)"></td>
        </tr>
      </table>
      <div class="info-and-controls">
        <div class="active-controls">
          <button v-if="selected.position != null" @click.stop.prevent="rotate">
            ROTATE (RIGHT CLICK)
          </button>
          <button v-if="selected.position != null" @click.stop.prevent="flip">
            FLIP (MOUSE WHEEL)
          </button>
          <button v-if="selected.position != null" @click.stop.prevent="place">
            PLACE (LEFT CLICK)
          </button>
        </div>
        <div class="guide">
          <button v-if="!showGuide" @click.stop.prevent="showGuide = !showGuide">
            SHOW GUIDE
          </button>
          <button v-if="showGuide" @click.stop.prevent="showGuide = !showGuide">
            HIDE GUIDE
          </button>
          <p v-if="showGuide">
            <ol>
              <li>FIRST PIECE OF A COLOR MUST COVER THAT COLORS STARTING CELL</li>
              <li>SAME COLOR PIECES CAN ONLY TOUCH AT THE CORNERS</li>
              <li>DIFFERENT COLORS CAN TOUCH FACES</li>
              <li>IF YOU CAN'T PLACE A PIECE PRESS THE END BUTTON</li>
              <li>THE WINNER IS THE PLAYER WITH THE LOWEST SCORE</li>
              <li>2 PLAYER - EACH PLAYER CONTROLS 2 COLORS</li>
              <li>3 PLAYER - EACH PLAYER CONTROLS 1 COLOR AND THE LAST COLOR IS CONTROLLED BY EACH PLAYER ON ROTATION</li>
              <li>4 PLAYER - EACH PLAYER CONTROLS 1 COLOR</li>
            </ol>
          </p>
        </div>
        <p v-if="gameIsActive() && game.myId == null">
          OBSERVING GAME, ENJOY :)
        </p>
        <div v-if="game.state === 0">
          <p>
            PLAYERS: {{game.players.length}}
          </p>
          <p v-if="game.players[0] !== game.myId">
            WAITING FOR MORE PLAYERS OR CREATOR TO START GAME
          </p>
          <p v-if="game.players[0] === game.myId">
            SEND LINK TO FRIENDS: <button @click.stop.prevent="copyLink">COPY LINK</button>
          </p>
          <button v-if="game.players[0] === game.myId && game.players.length >= 2" @click.stop.prevent="start">START GAME</button>
        </div>
        <p v-if="game.state >= 1">
          TURN: {{game.turn + 1}}
        </p>
        <p v-if="game.state === 2">
          GAME OVER
        </p>
        <p v-if="game.state === 2" class="winner" :class="'ps'+getWinningPlayerIdx()">
          P{{getWinningPlayerIdx()+1}} WON!
        </p>
        <p v-if="game.state === 3">
          GAME ABANDONED! SHAME! SHAME! SHAME!
        </p>
        <button v-if="game.state === 2 || game.state === 3" @click.stop.prevent="goToGames()">
          BACK TO GAMES
        </button>
      </div>
      <div class="piece-sets">
        <div v-for="(_, pieceSet) in pieceSetsCount" :key="pieceSet" :class="['piece-set',game.state === 1 && turnPieceSetIdx() === pieceSet?'active':'','ps'+pieceSet]">
          <div class="piece-set-header">
            <p v-if="game.state === 0" class="player-tag">
              WAITING TO START
            </p>
            <p v-if="game.state >= 1" class="player-tag">
              {{pieceSetPlayerLabel(pieceSet)}} <span v-if="game.players[turnPlayerIdx()] === game.myId && turnPieceSetIdx() === pieceSet">THAT'S YOU!</span>
            </p>
            <button :disabled="turnPieceSetIdx() !== pieceSet || game.players[turnPlayerIdx()] != game.myId || game.state !== 1" class="red" v-if="game.pieceSetsEnded[pieceSet] === '0'" @click.stop.prevent="end(pieceSet)">
              END
            </button>
            <p class="set-state" v-else>
              ENDED
            </p>
            <p class="set-score">
              SCORE: {{pieceSetScore(pieceSet)}}
            </p>
          </div>
          <div v-for="(_, piece) in pieces.length" :key="piece" class="piece">
            <table class="piece" :class="'ps'+pieceSet" @click.stop.prevent="select(pieceSet, piece)" v-if="game.pieceSets[(pieces.length*pieceSet)+piece] === '1'">
              <tr v-for="(_, y) in pieces[piece].bb[1]" :key="y">
                <td v-for="(_, x) in pieces[piece].bb[0]" :key="x" 
                  class="cell" :class="[pieceCell(piece, x, y)===1? 'ps'+pieceSet :'dead' ]"></td>
              </tr>
            </table>
          </div>
        </div>
      </div>
      <button class="red" v-if="gameIsActive() && game.myId != null" @click.stop.prevent="abandon">Abandon</button>
    </div>
  </div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'blockers',
    data: function() {
      window.x_api = api // todo delete this line
      return {
        pollInterval: 3000,
        gameType: "blockers",
        boardDims: 20,
        pieceSetsCount: 4,
        minPlayers: 2,
        loading: true,
        showGuide: false,
        myActiveGameRequested: false,
        errors: [],
        myActiveGame: {},
        game: {},
        selected: {},
        pieces: [
          // 0
          // #
          {bb: [1, 1], shape: [1], score: 1},

          // 1
          // ##
          {bb: [2, 1], shape: [1, 1], score: 2},

          // 2
          // ###
          {bb: [3, 1], shape: [1, 1, 1], score: 3},

          // 3
          // #
          // ##
          {bb: [2, 2], shape: [1, 0, 1, 1], score: 3},

          // 4
          // ####
          {bb: [4, 1], shape: [1, 1, 1, 1], score: 4},

          // 5
          // ##
          // ##
          {bb: [2, 2], shape: [1, 1, 1, 1], score: 4},

          // 6
          //  #
          // ###
          {bb: [3, 2], shape: [0, 1, 0, 1, 1, 1], score: 4},

          // 7
          //   #
          // ###
          {bb: [3, 2], shape: [0, 0, 1, 1, 1, 1], score: 4},

          // 8
          //  ##
          // ##
          {bb: [3, 2], shape: [0, 1, 1, 1, 1, 0], score: 4},

          // 9
          // #####
          {bb: [5, 1], shape: [1, 1, 1, 1, 1], score: 5},

          // 10
          // ###
          // ##
          {bb: [3, 2], shape: [1, 1, 1, 1, 1, 0], score: 5},

          // 11
          //  #
          // ###
          //  #
          {bb: [3, 3], shape: [0, 1, 0, 1, 1, 1, 0, 1, 0], score: 5},

          // 12
          // #
          // ###
          //   #
          {bb: [3, 3], shape: [1, 0, 0, 1, 1, 1, 0, 0, 1], score: 5},

          // 13
          //    #
          // ####
          {bb: [4, 2], shape: [0, 0, 0, 1, 1, 1, 1, 1], score: 5},

          // 14
          //   #
          // ####
          {bb: [4, 2], shape: [0, 0, 1, 0, 1, 1, 1, 1], score: 5},

          // 15
          // ###
          //   ##
          {bb: [4, 2], shape: [1, 1, 1, 0, 0, 0, 1, 1], score: 5},

          // 16
          // #
          // ###
          //  #
          {bb: [3, 3], shape: [1, 0, 0, 1, 1, 1, 0, 1, 0], score: 5},

          // 17
          // ###
          // # #
          {bb: [3, 2], shape: [1, 1, 1, 1, 0, 1], score: 5},

          // 18
          // #
          // ###
          // #
          {bb: [3, 3], shape: [1, 0, 0, 1, 1, 1, 1, 0, 0], score: 5},

          // 19
          // ##
          //  ##
          //   #
          {bb: [3, 3], shape: [1, 1, 0, 0, 1, 1, 0, 0, 1], score: 5},

          // 20
          // #
          // #
          // ###
          {bb: [3, 3], shape: [1, 0, 0, 1, 0, 0, 1, 1, 1], score: 5}
        ]
      }
    },
    methods: {
      link: function(){
        return window.location.href
      },
      copyLink: function(){
        let el = document.createElement('textarea');
        el.value = window.location.href;
        el.setAttribute('readonly', '');
        el.style = {position: 'absolute', left: '-9999px'};
        document.body.appendChild(el);
        el.select();
        document.execCommand('copy');
        document.body.removeChild(el);
      },
      get: function(){
        clearTimeout(this.getTimeoutId)
        let id = router.currentRoute.params.id
        let mapi = api
        let isMDoReq = false
        let promise = null
        if (!this.myActiveGameRequested) {
          isMDoReq = true
          mapi = api.newMDoApi()
          this.myActiveGameRequested = true
          mapi.game.active().then((info)=>{
            if (info != null) {
              this.myActiveGame = info
            }
          })
        }
        if (id === 'new') {
          promise = mapi.blockers.new().then((game)=>{
            this.game = game
            router.push('/'+this.gameType+'/'+this.game.id)
          }).catch((err)=>{
            let matches = err.body.match(/ id: ([^,]*), type: (.*)$/)
            if (matches.length === 3) {
              router.push('/'+matches[2]+'/'+matches[1])
            }
          })
        } else {
          let updatedAfter = null
          if (this.game != null && this.game.updatedOn != null) {
            updatedAfter = this.game.updatedOn
          }
          promise = mapi.blockers.get(id, updatedAfter).then((game)=>{
            if (game == null) {
              // if we asked for data only after updatedAfter
              // null will be returned if there's no updated state.
              return
            }
            // reset selected
            this.selected = {}
            this.game = game
          }).catch((err)=>{
            if (err.status === 404) {
              router.push('/games')
            }
          })
        }
        if (isMDoReq) {
          promise = mapi.sendMDo()
        }
        promise.catch((error)=>{
          if (error.isMDoErrors === true) {
            for (let i = 0, l = error.length; i < l; i++) {
              this.errors.push(error[i])
            }
          } else {
            this.errors.push(error)
          }
        }).finally(()=>{
          let poll = ()=>{
            this.loading = false
            if (this.game.id != null && this.gameIsActive()) {
              this.getTimeoutId = setTimeout(this.get, this.pollInterval)
            }
          }
          if (this.game.state === 0 && 
            this.game.myId == null && 
            this.game.players.length < this.pieceSetsCount &&
            this.myActiveGame.id == null) {
              api.blockers.join(this.game.id).then((game)=>{
                this.game = game
                poll()
            })
          } else {
            poll()
          }
        })
      },
      select: function(pieceSet, piece) {
        if (this.game != null &&
          this.game.state === 1 &&
          this.game.players[this.turnPlayerIdx()] === this.game.myId &&
          this.turnPieceSetIdx() === pieceSet &&
          this.game.pieceSets[(this.pieces.length*pieceSet)+piece] === '1') {
          let bb = this.pieces[piece].bb
          let bbCpy = [bb[0], bb[1]]
          let shape = this.pieces[piece].shape
          let shapeCpy = []
          for (let i = 0, l = shape.length; i < l; i++) {
            shapeCpy.push(shape[i])
          }
          // place it in the center of the board by default
          let x = Math.floor(this.boardDims / 2) - Math.floor(bb[0] / 2)
          let y = Math.floor(this.boardDims / 2) - Math.floor(bb[1] / 2)

          this.selected = {
            position: this.xyToI(x, y, this.boardDims),
            flip: 0,
            rotation: 0,
            piece: piece,
            pieceSet: pieceSet,
            bb: bbCpy,
            shape: shapeCpy,
            activeCellConMet: true,
            firstCornerConMet: false,
            cornerConMet: false,
            faceConMet: true,
            activeCells: {},
          }
          this.updateSelectedActiveCells()
          window.scrollTo({top: 0, left: 0, behavior: "smooth"})
        }
      },
      updateSelectedActiveCells: function(){
        if (this.selected.position != null) {
          // cpy shape and bb fresh as flip and rotate need to be applied
          // in specific order
          let bb = this.pieces[this.selected.piece].bb
          let bbCpy = [bb[0], bb[1]]
          let shape = this.pieces[this.selected.piece].shape
          let shapeCpy = []
          for (let i = 0, l = shape.length; i < l; i++) {
            shapeCpy.push(shape[i])
          }
          this.selected.bb = bbCpy
          this.selected.shape = shapeCpy
          // flip 
          if (this.selected.flip === 1) {
            let flippedShape = []
            for ( let y = 0; y < this.selected.bb[1]; y++) {
              for (let x = 0; x < this.selected.bb[0]; x++) {
                flippedShape[(y*this.selected.bb[0])+x] = this.selected.shape[(y*this.selected.bb[0])+this.selected.bb[0]-1-x]
              }
            }
            this.selected.shape = flippedShape
          }
          // rotate
          for (let i = 0; i < this.selected.rotation; i++) {
            let rotatedShape = []
            for (let y = 0; y < this.selected.bb[1]; y++) {
              for (let x = 0; x < this.selected.bb[0]; x++) {
                rotatedShape[(x*this.selected.bb[1])+(this.selected.bb[1]-1-y)] = this.selected.shape[(y*this.selected.bb[0])+x]
              }
            }
            this.selected.shape = rotatedShape
            let bb0 = this.selected.bb[0]
            this.selected.bb[0] = this.selected.bb[1]
            this.selected.bb[1] = bb0
          }
          this.selected.activeCells = new Map()
          let piecePos = this.iToXY(this.selected.position, this.boardDims, this.boardDims)
          for (let i = 0, l = this.selected.shape.length; i < l; i++) {
            if (this.selected.shape[i] === 1) {
              // get cells coords on board
              let cellPos = this.iToXY(i, this.selected.bb[0], this.selected.bb[1])
              let cellX = cellPos.x + piecePos.x
              let cellY = cellPos.y + piecePos.y
              for (let offsetY = -1; offsetY < 2; offsetY++) {
                for (let offsetX = -1; offsetX < 2; offsetX++) {
                  let loopX = cellX + offsetX
                  let loopY = cellY + offsetY
                  if (loopX >= 0 &&
                  loopY >= 0 &&
                  loopX < this.boardDims &&
                  loopY < this.boardDims) {
                    // set to 'c'orner, 'f'ace, 'a'tive cell
                    let boardI = this.xyToI(loopX, loopY, this.boardDims)
                    if (offsetX === 0 &&
                    offsetY === 0) {
                      this.selected.activeCells.set(boardI, 'a')
                    } else if ((offsetX === 0 || 
                    offsetY === 0) && 
                    this.selected.activeCells.get(boardI) !== 'a') {
                      this.selected.activeCells.set(boardI, 'f')
                    } else if (this.selected.activeCells.get(boardI) == null) {
                      this.selected.activeCells.set(boardI, 'c')
                    }
                  }
                }
              }
            }
          }
          // loop over finalized active cells and check
          // if constraints are met.
          let startI = this.pieceSetBoardStartI(this.selected.pieceSet)
          this.selected.firstCornerConMet =
            this.game.board[startI] !== '4'
          this.selected.cornerConMet = !this.selected.firstCornerConMet
          this.selected.activeCellConMet = true
          this.selected.faceConMet = true
          this.selected.activeCells.forEach((v, i)=>{
            this.selected.firstCornerConMet = this.selected.firstCornerConMet || (i === startI && v === 'a')
            this.selected.cornerConMet = this.selected.cornerConMet || (v === 'c' && this.game.board[i] === ''+this.selected.pieceSet)
            if (v === 'a') {
              this.selected.activeCellConMet = this.selected.activeCellConMet && this.game.board[i] === '4'
            }
            if (v === 'f') {
              this.selected.faceConMet = this.selected.faceConMet && this.game.board[i] !== ''+this.selected.pieceSet
            }
          })
        }
      },
      boardCellClass: function(x, y) {
        let i = this.xyToI(x, y, this.boardDims)
        if (this.selected.piece != null && 
        this.selected.activeCells.get(i) != null) {
          let cellState = this.selected.activeCells.get(i)
          switch (cellState) {
            // 'a'ctive
            case 'a':
              if (this.game.board[i] !== '4') {
                return 'ps'+this.selected.pieceSet+' shade'
              }
              return 'ps'+this.selected.pieceSet
            // 'f'ace
            case 'f':
              if (this.selected.activeCellConMet && this.game.board[i] === ''+this.selected.pieceSet) {
                return 'ps'+this.selected.pieceSet+' shade'
              }
              break;
            // 'c'orner
            case 'c':
              if (this.selected.activeCellConMet && this.selected.faceConMet && !this.selected.cornerConMet) {
                return 'ps'+this.selected.pieceSet+' shade'
              }
              break;
          }
        }

        if (this.game.board[i] === '4') {
          if (this.selected.position == null ||
          i === this.pieceSetBoardStartI(this.selected.pieceSet)) {
            // only apply start cell styles to start cells
            // with no pieces over the top of them.
            if (i === this.pieceSetBoardStartI(0)) {
              return "ps0 shade"
            } else if (i === this.pieceSetBoardStartI(1)) {
              return "ps1 shade"
            } else if (i === this.pieceSetBoardStartI(2)) {
              return "ps2 shade"
            } else if (i === this.pieceSetBoardStartI(3)) {
              return "ps3 shade"
            }
          }
        }
        return "ps"+this.game.board[i]
      },
      place: function() {
        if (this.selected.position != null &&
        this.selected.firstCornerConMet &&
        this.selected.activeCellConMet &&
        this.selected.faceConMet &&
        this.selected.cornerConMet) {
          clearTimeout(this.getTimeoutId)
          api.blockers.takeTurn(
            this.selected.piece,
            this.selected.position,
            this.selected.flip,
            this.selected.rotation,
            0).then((game)=>{
              this.game = game
              this.getTimeoutId = setTimeout(this.get, this.pollInterval)
            }).finally(()=>{
              this.selected = {}
            })
        }
      },
      rotate: function() {
        if (this.selected.position != null) {
          this.selected.rotation++
          this.selected.rotation %= 4
          let pos = this.iToXY(this.selected.position, this.boardDims)
          let offsetX = this.boardDims - this.selected.bb[1] - pos.x
          let offsetY = this.boardDims - this.selected.bb[0] - pos.y
          if (offsetX < 0) {
            pos.x += offsetX
          }
          if (offsetY < 0) {
            pos.y += offsetY
          }
          this.selected.position = this.xyToI(pos.x, pos.y, this.boardDims)
          this.updateSelectedActiveCells()
        }
      },
      flip: function(event) {
        if (this.selected.position != null) {
          event.stopPropagation()
          event.preventDefault()
          if (this.selected.flip === 0) {
            this.selected.flip = 1
          } else {
            this.selected.flip = 0
          }
          this.updateSelectedActiveCells()
        }
      },
      move: function(x, y) {
        if (this.game != null &&
          this.game.state === 1 &&
          this.selected.position != null) {
          let offsetX = this.boardDims - this.selected.bb[0] - x
          let offsetY = this.boardDims - this.selected.bb[1] - y
          if (offsetX < 0) {
            x += offsetX
          }
          if (offsetY < 0) {
            y += offsetY
          }
          this.selected.position = this.xyToI(x, y, this.boardDims)
          this.updateSelectedActiveCells()
        }
      },
      touchBoard: function(event){
        if (event != null && 
        event.target != null &&
        event.target.dataset != null &&
        event.target.dataset.x != null &&
        event.target.dataset.y != null &&
        this.selected.position != null &&
        event.touches.length === 1) {
          event.stopPropagation()
          event.preventDefault()
          let changedTouch = event.changedTouches[0];
          let target = document.elementFromPoint(changedTouch.clientX, changedTouch.clientY);
          if (target.dataset.x != null &&
          target.dataset.y != null) {
            this.move(parseInt(target.dataset.x), parseInt(target.dataset.y))
          }
        }
      },
      turnPlayerIdx: function(){
        let playerIdx = 0
        if (this.game.players.length === 3) {
          playerIdx = this.game.turn % this.pieceSetsCount
          if (playerIdx == 3) {
            playerIdx = (Math.floor(((this.game.turn) + 1) / this.pieceSetsCount) - 1) % 3
          }
        } else {
          playerIdx = this.game.turn % this.game.players.length
        }
        return playerIdx
      },
      turnPieceSetIdx: function(){
        return this.game.turn % this.pieceSetsCount
      },
      pieceSetPlayerLabel: function(pieceSet){
        switch (pieceSet) {
          case 0:
          case 1:
            return 'P'+(pieceSet+1)
          case 2:
            switch (this.game.players.length) {
              case 2:
                return 'P1'
              default:
                return 'P3'
            }
          case 3:
            switch (this.game.players.length) {
              case 2:
                return 'P2'
              case 4:
                return 'P4'
              case 3:
                if (this.turnPieceSetIdx() !== 3) {
                  return 'ROTA'
                }
                return 'P'+ (this.turnPlayerIdx()+1);
            }
        }
      },
      pieceSetScore: function(pieceSet){
        let score = 0
        for (let i = 0, l = this.pieces.length; i < l; i++) {
          if (this.game.pieceSets[(this.pieces.length*pieceSet)+i] === '1') {
            score += this.pieces[i].score
          }
        }
        return score
      },
      getWinningPlayerIdx: function() {
        let scores = []
        switch (this.game.players.length) {
          case 2:
            scores = [this.pieceSetScore(0)+this.pieceSetScore(2), this.pieceSetScore(1)+this.pieceSetScore(3)]
            break
          case 3:
            scores = [this.pieceSetScore(0), this.pieceSetScore(1), this.pieceSetScore(2)]
            break
          case 4:
            scores = [this.pieceSetScore(0), this.pieceSetScore(1), this.pieceSetScore(2), this.pieceSetScore(3)]
        }
        let winningPlayer = 0
        for (let i = 1, l = scores.length; i < l; i++) {
          if (scores[winningPlayer] >= scores[i]) {
            winningPlayer = i
          }
        }
        return winningPlayer
      },
      goToMyActiveGame: function(){
        if (this.myActiveGame.id != null) {
          router.push('/'+this.myActiveGame.type+'/'+this.myActiveGame.id)
        }
      },
      goToGames: function(){
        router.push('/games')
      },
      xyToI: function(x, y, xDim){
        return xDim*y + x
      },
      iToXY: function(i, xDim){
        let x = i % xDim
        let y = Math.floor((i-x) / xDim)
        return {x, y}
      },
      pieceCell: function(piece, x, y){
        let p = this.pieces[piece]
        return p.shape[this.xyToI(x, y, p.bb[0])]
      },
      pieceSetBoardStartI: function(pieceSet) {
        switch (pieceSet) {
          case 0:
            return this.xyToI(0, 0, this.boardDims)
          case 1:
            return this.xyToI(this.boardDims-1, 0, this.boardDims)
          case 2:
            return this.xyToI(this.boardDims-1, this.boardDims-1, this.boardDims)
          case 3:
            return this.xyToI(0, this.boardDims-1, this.boardDims)
        }
      },
      gameIsActive: function(){
        return this.game.state === 0 || this.game.state === 1
      },
      start: function(){
        api.blockers.start(false).then((game)=>{
          this.game = game
        })
      },
      end: function(pieceSet){
        if (this.turnPieceSetIdx() === pieceSet && this.game.players[this.turnPlayerIdx()] === this.game.myId) {
          this.selected = {}
          api.blockers.takeTurn(0, 0, 0, 0, 1).then((game)=>{
            this.game = game
          })
        }
      },
      abandon: function(){
        if (window.confirm("Do you really want to abandon this game?!?!")) {
          clearTimeout(this.getTimeoutId)
          api.blockers.abandon().then(()=>{
            this.game.state = 3
          })
        }
      }
    },
    created: function(){
      this.get()
      // add event listeners
    },
    destroyed: function(){
      clearTimeout(this.getTimeoutId)
      // remove event listeners
    },
    watch: {
      $route () {
        this.get()
      }
    },
  }
</script>

<style scoped lang="scss">
@use "sass:color";

$ps0: #d33;
$ps1: #3d3;
$ps2: #33d;
$ps3: #dd3;
$empty: #222;
$border: 1px solid black;
$cellSize: 2pc;

@mixin shade($base) {
  background: repeating-linear-gradient(
      45deg,
      color.adjust($base, $lightness: -26),
      color.adjust($base, $lightness: -26) 5px,
      color.adjust($base, $lightness: -15) 5px,
      color.adjust($base, $lightness: -15) 10px
    );
}
@mixin cell($root, $base) {
  &.#{$root}{
    color: white; //todo set this back to $base
    background: $base;
    &.dead{
      background: transparent;
    }
    &.shade{
      @include shade($base)
    }
  }
}
@mixin pieceSet($root, $base) {
  &.#{$root}{
    @include shade($base);
    &:not(.active){
      @include shade(color.adjust($base, $lightness: -27));
      .cell {
        background: color.adjust($base, $lightness: -36);
        &.dead{
          background: transparent;
        }
      }
    }
    &.active{
      .piece-set-header{
      @include shade(#777);
        p {
          color: $base;
        }
      }
    }
  }
}

.game > *{
  margin: 1pc;
}

table, tr, td{
  border-collapse: collapse;
  background: transparent;
}

.board {
  min-width: 44pc;
}

.cell{
  cursor: pointer;
  width: $cellSize;
  height: $cellSize;
  &:not(.dead){
    border: $border
  }
  @include cell("ps0", $ps0);
  @include cell("ps1", $ps1);
  @include cell("ps2", $ps2);
  @include cell("ps3", $ps3);
  @include cell("ps4", $empty);
}

.piece-sets {
    white-space: nowrap;
  > .piece-set {
    vertical-align: top;
    display: inline-block;
    border: 1px solid #555;
    @include pieceSet('ps0', $ps0);
    @include pieceSet('ps1', $ps1);
    @include pieceSet('ps2', $ps2);
    @include pieceSet('ps3', $ps3);

    > .piece-set-header{
      font-weight: 900;
      align-content: center;
      @include shade(#333);
      p {
        margin: 10px 0;
        color: #555;
        background: transparent;
        text-align: center;
        > span {
          font-weight: inherit;
          color: inherit;
          background: transparent;
        }
      }
      button.red{
        position: relative;
        width: 5pc;
        left: calc(50% - 2.5pc);
        &:disabled, &[disabled]{
          background: #100;
          color: #555;
        }
      }
    }
    > .piece {
      background: transparent;
      float: left;
      clear: both;
      > table {
        margin: 0.5pc;
      }
    }
    &.active > .piece:hover td:not(.dead){
      border: 1px solid white;      
    }
  }
}

.info-and-controls{
  .active-controls button{
    margin: 0pc 1pc 1pc 0pc;
  }
  .winner{
    &.ps0{
      color: $ps0;
    }
    &.ps1{
      color: $ps1;
    }
    &.ps2{
      color: $ps2;
    }
    &.ps3{
      color: $ps3;
    }
  }
}

button.red{
  background: #500;
  border: 1px solid #300;
  &:hover {
      background-color: #700;
  }
  &:active {
      background-color: #500;
  }
}
</style>