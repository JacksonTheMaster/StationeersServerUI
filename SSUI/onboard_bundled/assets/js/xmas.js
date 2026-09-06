document.addEventListener('DOMContentLoaded', function() {

  // Check if the date is between Dec 1 and Jan 10 (inclusive, across year boundary)
  const now = new Date();
  const isXmas = (now.getMonth() === 11 && now.getDate() >= 1) ||  // Dec 1–31
                 (now.getMonth() === 0 && now.getDate() <= 10);    // Jan 1–10

  // If Christmas period is over, do nothing
  if (!isXmas) {
    return;
  }

  console.log('Christmas sleigh ready to fly');

  // Inject the sleigh CSS
  const sleighStyle = document.createElement('style');
  sleighStyle.textContent = `
    .sleigh-container {
      position: fixed;
      z-index: -1;
      transition: transform 0.3s;
      animation: flySleigh 30s linear infinite;
      scale: 0.5;
    }
    
    .sleigh-container:hover {
      transform: scale(1.1);F
    }
    
    @keyframes flySleigh {
      0% {
        left: -400px;
        top: 40%;
      }
      50% {
        top: 30%;
      }
      100% {
        left: calc(100% + 400px);
        top: 20%;
    }
}
    
    @keyframes bounce {
      0%, 100% { transform: translateX(-50%) translateY(0); }
      50% { transform: translateX(-50%) translateY(-8px); }
    }

    /* Sleigh styles */
    .sleigh-santa {
      position: relative;
      width: 295px;
      height: 155px;
      transform: rotate(-1deg);
    }

    .santa {
      position: absolute;
      bottom: 0;
      left: 50%;
      width: 125px;
      height: 107px;
      z-index: 10;
    }

    .santa--sleigh {
      bottom: 0;
      left: 0;
      transform: rotateY(180deg);
    }

    .santa--sleigh:before,
    .santa--sleigh:after {
      content: "";
      position: absolute;
      bottom: 0;
      background-color: #8B0000;
    }

    .santa--sleigh:before {
      left: -10px;
      width: 129px;
      height: 30px;
      border-radius: 5px 5% 10px 65%;
      transform: rotate(0);
      z-index: 10;
      border-bottom: 2px solid #DAA520;
    }

    .santa--sleigh:after {
      border: 2px solid #DAA520;
      left: 70px;
      bottom: 0px;
      width: 50px;
      height: 57px;
      border-radius: 50% 10px 16px 10px;
      transform: rotate(1deg);
      box-shadow: -98px -2px 0px -18px #8B0000;
    }

    .santa__hat-part {
      position: absolute;
      top: 7px;
      left: 31px;
      width: 43px;
      height: 58px;
      border-radius: 50%;
      transform: rotate(28deg);
      background-color: #d63527;
    }

    .santa__hat-part:before,
    .santa__hat-part:after {
      content: "";
      position: absolute;
    }

    .santa__hat-part:nth-of-type(1):before {
      top: 9px;
      left: 45px;
      width: 7px;
      height: 7px;
      border-radius: 50%;
      background-color: #fff;
      animation: santa-hat-bobble 1s linear alternate infinite;
    }

    .santa__hat-part:nth-of-type(1):after {
      top: 3px;
      left: 19px;
      width: 30px;
      height: 7px;
      border-radius: 50%;
      transform: rotate(22deg);
      background-color: #d63527;
      animation: santa-hat-main 1s linear alternate infinite;
    }

    .santa__hat-part:nth-of-type(2) {
      position: absolute;
      top: 18px;
      left: 31px;
      width: 44px;
      height: 34px;
      border-radius: 50%;
      transform: rotate(12deg);
      background-color: #fff;
    }

    .santa__face {
      position: absolute;
      top: 25px;
      left: 37px;
      width: 31px;
      height: 17px;
      border-radius: 20px 20px 50% 50%;
      transform: rotate(10deg);
      background-color: #fde2b7;
      z-index: 10;
    }

    .santa__beard-part {
      position: absolute;
      top: 8px;
      left: -14px;
      width: 15px;
      height: 17px;
      border-radius: 50%;
      background-color: #fff;
    }

    .santa__beard-part:before,
    .santa__beard-part:after {
      content: "";
      position: absolute;
      background-color: #fff;
    }

    .santa__beard-part:before {
      top: 12px;
      left: 1px;
      width: 15px;
      height: 17px;
      border-radius: 50%;
    }

    .santa__beard-part:nth-of-type(2) {
      top: 16px;
      left: -8px;
      width: 26px;
      height: 30px;
    }

    .santa__beard-part:nth-of-type(2):before {
      top: 16px;
      left: 13px;
      width: 19px;
      height: 17px;
    }

    .santa__beard-part:nth-of-type(2):after {
      top: 1px;
      left: 13px;
      width: 19px;
      height: 17px;
    }

    .santa__beard-part:nth-of-type(3) {
      top: 16px;
      left: 14px;
      width: 27px;
      height: 28px;
    }

    .santa__beard-part:nth-of-type(3):before {
      top: -4px;
      left: 13px;
      width: 17px;
      height: 17px;
    }

    .santa__eyebrows {
      position: absolute;
      top: 0;
      left: 0;
      width: 2px;
      height: 7px;
      border-radius: 50%;
      background-color: #fff;
    }

    .santa__eyebrows--left {
      top: 1px;
      left: 4px;
      transform: rotate(65deg);
    }

    .santa__eyebrows--right {
      top: 2px;
      left: 22px;
      transform: rotate(-65deg);
    }

    .santa__eye {
      position: absolute;
      width: 3px;
      height: 4px;
      border-radius: 50%;
      background-color: #000;
    }

    .santa__eye--left {
      top: 8px;
      left: 2px;
    }

    .santa__eye--right {
      top: 8px;
      left: 20px;
    }

    .santa__nose {
      position: absolute;
      top: 10px;
      left: 6px;
      width: 12px;
      height: 9px;
      border-radius: 50%;
      z-index: 10;
      background-color: #f7d194;
    }

    .santa__cheek {
      position: absolute;
      width: 7px;
      height: 7px;
      border-radius: 50%;
      z-index: 10;
      background-color: #f4cfe3;
    }

    .santa__cheek--left {
      top: 12px;
      left: -3px;
    }

    .santa__cheek--right {
      top: 13px;
      left: 22px;
    }

    .santa__body {
      position: absolute;
      top: 54px;
      left: 16px;
      width: 88px;
      height: 53px;
    }

    .santa__body:before {
      content: "";
      position: absolute;
      top: -23px;
      right: -10px;
      width: 53px;
      height: 51px;
      border-radius: 42% 50%;
      background-color: #362312;
      z-index: -1;
      box-shadow: 10px -21px 0px -20px #e1b12c, 15px -30px 0px -18px #362312;
      animation: santa-sac 0.6s linear alternate infinite;
    }

    .santa__body-top {
      top: -3px;
      left: 10px;
      position: absolute;
      width: 45px;
      height: 39px;
      border-radius: 50% 50% 10% 10%;
      background-color: #d63527;
      z-index: 5;
    }

    .santa__body-top:before {
      content: "";
      top: 28px;
      left: 0px;
      position: absolute;
      width: 45px;
      height: 5px;
      background-color: #000;
      transform: rotate(1deg);
    }

    .santa__body-top:after {
      content: "";
      top: 27px;
      left: 10px;
      position: absolute;
      width: 7px;
      height: 5px;
      background-color: #000;
      border: 1px solid #fff;
      border-radius: 3px;
      transform: rotate(1deg);
    }

    .santa__hand--left {
      top: 5px;
      left: 19px;
      width: 33px;
      height: 30px;
      overflow: hidden;
      position: absolute;
    }

    .santa__hand-inner {
      position: absolute;
      top: 10px;
      left: 8px;
      width: 49px;
      z-index: 100;
      height: 7px;
      border-radius: 10px;
      transform: rotate(12deg);
      background-color: #d63527;
      animation: sleigh-santa-hand-left 1s linear alternate infinite;
    }

    .santa__hand-inner:before {
      content: "";
      position: absolute;
      width: 8px;
      height: 7px;
      top: -2px;
      left: -6px;
      background-color: #000;
      border-radius: 50%;
      transform: rotate(25deg);
    }

    .santa__hand--right {
      position: absolute;
      top: 4px;
      left: 3px;
      width: 11px;
      height: 8px;
      transform: rotate(25deg);
      border-radius: 10px;
      height: 7px;
      background-color: #d63527;
      animation: sleigh-santa-hand-right 1s linear alternate infinite;
    }

    .santa__hand--right:before {
      content: "";
      position: absolute;
      width: 8px;
      height: 7px;
      top: -2px;
      left: -6px;
      background-color: #000;
      border-radius: 50%;
      transform: rotate(10deg);
    }

    .sleigh-feet {
      position: absolute;
      bottom: -10px;
      left: 0px;
      width: 145px;
      height: 11px;
      transform: rotate(-5deg);
      border-bottom: 5px solid #996515;
      border-right: 5px solid #996515;
      border-radius: 10px;
      z-index: 2;
    }

    .sleigh-feet:before,
    .sleigh-feet:after {
      content: "";
      position: absolute;
      top: 0;
      left: 0;
      width: 5px;
      height: 5px;
      background-color: #996515;
    }

    .sleigh-feet:before {
      top: 2px;
      left: 34px;
      height: 9px;
    }

    .sleigh-feet:after {
      top: 3px;
      left: 108px;
      width: 5px;
      height: 8px;
    }

    .lead {
      position: absolute;
      top: 92px;
      left: 84px;
      width: 182px;
      height: 33px;
      overflow: hidden;
      z-index: 10;
      transform: rotate(0deg);
      animation: sleigh-santa-lead-right 1s linear alternate infinite;
    }

    .lead--back {
      top: 85px;
      left: 105px;
      width: 149px;
      transform: rotate(4deg);
      z-index: 0;
      animation: sleigh-santa-lead-left 1s linear alternate infinite;
    }

    .lead-inner {
      position: absolute;
      bottom: 0;
      left: -12px;
      width: 100%;
      height: 48px;
      border-bottom: 1px solid #fff;
      border-radius: 50%;
    }

    .reindeer {
      position: absolute;
      width: 115px;
      height: 155px;
      right: 5px;
      top: 17px;
      transform: rotate(25deg);
      z-index: 0;
    }

    .reindeer:before {
      content: "";
      position: absolute;
      top: 65px;
      left: 76px;
      width: 8px;
      height: 31px;
      background-color: #8B0000;
      z-index: 10;
      transform: rotate(-55deg);
    }

    .reindeer:after {
      content: "";
      position: absolute;
      height: 5px;
      width: 5px;
      border-radius: 50%;
      background-color: #DAA520;
      z-index: 11;
      left: 64px;
      top: 68px;
      box-shadow: 8px 6px 0 0 #DAA520, 18px 13px 0 0 #DAA520, 26px 18px 0 0 #DAA520;
    }

    .reindeer__face {
      position: absolute;
      width: 30px;
      height: 22px;
      top: 44px;
      left: 72px;
      border-radius: 10px 10px 50% 50%;
      transform: rotate(-3deg);
      background-color: #654321;
      animation: reindeer-face 1.6s linear alternate infinite;
    }

    .reindeer__face:before {
      content: "";
      position: absolute;
      background-color: #654321;
      width: 29px;
      height: 16px;
      border-radius: 50%;
      top: 0px;
      left: 11px;
      transform: rotate(-49deg);
    }

    .reindeer__face:after {
      content: "";
      position: absolute;
      background-color: #8B0000;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      top: -8px;
      left: 31px;
    }

    .reindeer__horn {
      position: absolute;
      width: 29px;
      height: 4px;
      border-radius: 2px;
      background-color: #DEB887;
    }

    .reindeer__horn:before,
    .reindeer__horn:after {
      content: "";
      position: absolute;
      background-color: #DEB887;
      border-radius: 2px;
    }

    .reindeer__horn--left {
      top: -7px;
      left: -21px;
      transform: rotate(38deg);
    }

    .reindeer__horn--left:before {
      top: -4px;
      left: 6px;
      width: 14px;
      height: 4px;
      transform: rotate(43deg);
    }

    .reindeer__horn--left:after {
      top: -4px;
      left: 13px;
      width: 14px;
      height: 4px;
      transform: rotate(53deg);
    }

    .reindeer__horn--right {
      top: -12px;
      left: -6px;
      width: 24px;
      transform: rotate(62deg);
    }

    .reindeer__horn--right:before {
      top: -3px;
      left: 5px;
      width: 10px;
      height: 4px;
      transform: rotate(43deg);
    }

    .reindeer__horn--right:after {
      top: -3px;
      left: 11px;
      width: 10px;
      height: 4px;
      transform: rotate(53deg);
    }

    .reindeer__ear {
      position: absolute;
      width: 21px;
      height: 11px;
      top: 4px;
      left: -18px;
      border-radius: 4px 0 50% 50%;
      transform: rotate(4deg);
      background-color: #654321;
    }

    .reindeer__ear:before {
      content: "";
      position: absolute;
      top: -2px;
      left: 34px;
      width: 4px;
      height: 5px;
      border-radius: 50%;
      transform: rotate(-35deg);
      background-color: #000;
    }

    .reindeer__body {
      position: absolute;
      width: 58px;
      height: 31px;
      top: 84px;
      left: 28px;
      border-radius: 50% 0;
      transform: rotate(-3deg);
      background-color: #654321;
    }

    .reindeer__body:before {
      content: "";
      position: absolute;
      width: 46px;
      height: 26px;
      top: -15px;
      left: 32px;
      border-radius: 0 0 50% 50%;
      transform: rotate(-55deg);
      background-color: #654321;
    }

    .reindeer__body:after {
      content: "";
      position: absolute;
      width: 43px;
      height: 26px;
      top: -11px;
      left: 29px;
      border-radius: 0 0 50% 50%;
      transform: rotate(-30deg);
      background-color: #654321;
    }

    .reindeer__foot--inside {
      z-index: 2;
      transform: rotate(-12deg) translate(3px, 0px);
    }

    .reindeer__foot-inner {
      position: absolute;
    }

    .reindeer__foot-inner:before,
    .reindeer__foot-inner:after {
      content: "";
      position: absolute;
    }

    .reindeer__foot--front .reindeer__foot-inner {
      width: 40px;
      height: 8px;
      top: 13px;
      left: 35px;
      border-radius: 0 50%;
      transform: rotate(-17deg);
      transform-origin: center;
      background-color: #654321;
      animation: reindeer-front 1.6s linear alternate infinite;
    }

    .reindeer__foot--front .reindeer__foot-inner:before {
      width: 28px;
      height: 8px;
      top: 0px;
      left: 37px;
      border-radius: 2px 50%;
      transform: rotate(131deg);
      background-color: #654321;
      animation: reindeer-front-ext 1.7s linear alternate infinite;
    }

    .reindeer__foot--front .reindeer__foot-inner:after {
      width: 8px;
      height: 9px;
      top: 27px;
      left: 32px;
      border-radius: 2px;
      transform: rotate(131deg);
      background-color: #362514;
      animation: reindeer-front-ext-hoof 1.7s linear alternate infinite;
    }

    .reindeer__foot--back .reindeer__foot-inner {
      width: 56px;
      height: 9px;
      top: 35px;
      left: -29px;
      border-radius: 0 50%;
      transform: rotate(-73deg);
      background-color: #654321;
      animation: reindeer-back 1.7s linear alternate infinite;
    }

    .reindeer__foot--back .reindeer__foot-inner:before {
      width: 25px;
      height: 16px;
      top: 4px;
      left: 25px;
      border-radius: 0 50%;
      transform: rotate(15deg);
      background-color: #654321;
    }

    .reindeer__foot--back .reindeer__foot-inner:after {
      width: 8px;
      height: 9px;
      top: -2px;
      left: -2px;
      border-radius: 2px 0 2px 2px;
      transform: rotate(14deg);
      background-color: #362514;
    }

    .reindeer__tail {
      position: absolute;
      width: 27px;
      height: 26px;
      top: 6px;
      left: -8px;
      border-radius: 50% 2px;
      transform: rotate(-17deg);
      background-color: #654321;
    }

    .reindeer__tail:before {
      content: "";
      position: absolute;
      background-color: #654321;
      border-radius: 50%;
      top: -2px;
      left: -3px;
      width: 15px;
      height: 5px;
      transform: rotate(25deg);
    }

    .reindeer__spots {
      position: absolute;
      width: 4px;
      height: 4px;
      top: 6px;
      left: 8px;
      border-radius: 50% 2px;
      background-color: #DEB887;
      box-shadow: 5px 5px 0 0 #DEB887, -5px 5px 0 0 #DEB887;
    }

    .particles {
      position: absolute;
      bottom: -30px;
      left: 0;
      width: 7px;
      height: 7px;
      background-color: transparent;
      transform: rotate(45deg);
      z-index: -10;
    }

    .particles:before,
    .particles:after {
      position: absolute;
      content: "";
      background-color: #FFF;
    }

    .particles:after {
      left: 0px;
      top: 0px;
      width: 4px;
      height: 4px;
      transform: rotate(-5deg);
      animation: particles 4s linear infinite;
      box-shadow: -20px 15px 0px 0px #FFF, -40px -5px 0px 0px #FFF, -20px 45px 0px 0px #FFF, -50px 30px 0px 0px #FFF, 30px -20px 0px 0px #FFF, 50px -60px 0px 0px #FFF, 100px -110px 0px 0px #FFF, 140px -160px 0px 0px #FFF, 50px -90px 0px 0px #FFF, 100px -140px 0px 0px #FFF, 140px -190px 0px 0px #FFF;
    }

    .particles:before {
      left: 10px;
      top: 0px;
      width: 2px;
      height: 2px;
      transform: rotate(-10deg);
      animation: particles 5s linear infinite;
      box-shadow: -20px 15px 0px 0px #FFF, -40px -5px 0px 0px #FFF, -20px 45px 0px 0px #FFF, -50px 30px 0px 0px #FFF, 30px -20px 0px 0px #FFF, 50px -60px 0px 0px #FFF, 100px -110px 0px 0px #FFF, 140px -160px 0px 0px #FFF, 50px -90px 0px 0px #FFF, 100px -140px 0px 0px #FFF, 140px -190px 0px 0px #FFF;
    }

    @keyframes reindeer-face {
      0% { transform: rotate(-8deg); }
      100% { transform: rotate(2deg); }
    }

    @keyframes reindeer-back {
      0% { transform: rotate(-81deg) translate(-6px, 0px); }
      100% { transform: rotate(-60deg) translate(0px, 0px); }
    }

    @keyframes reindeer-front {
      0% { transform: rotate(-24deg); }
      100% { transform: rotate(-13deg); }
    }

    @keyframes reindeer-front-ext {
      0% {
        transform-origin: left;
        transform: rotate(55deg);
      }
      100% {
        transform-origin: left;
        transform: rotate(145deg);
      }
    }

    @keyframes reindeer-front-ext-hoof {
      0% {
        transform-origin: left top;
        transform: rotate(-32deg) translate(14px, 1px);
      }
      100% {
        transform-origin: left top;
        transform: rotate(52deg) translate(-21px, 0px);
      }
    }

    @keyframes sleigh-santa-lead-left {
      0% { transform: rotate(1deg); }
      100% { transform: rotate(-2deg); }
    }

    @keyframes sleigh-santa-hand-left {
      0% { transform: rotate(15deg); }
      100% { transform: rotate(-2deg); }
    }

    @keyframes sleigh-santa-lead-right {
      0% { transform: rotate(2deg); }
      100% { transform: rotate(-4deg); }
    }

    @keyframes sleigh-santa-hand-right {
      0% { transform: rotate(15deg); }
      100% { transform: rotate(-30deg); }
    }

    @keyframes santa-hat-main {
      0% {
        transform-origin: left;
        transform: rotate(-15deg);
      }
      100% {
        transform-origin: left;
        transform: rotate(15deg);
      }
    }

    @keyframes santa-hat-bobble {
      0% { transform: translate(0px, -15px); }
      100% { transform: translate(0px, 5px); }
    }

    @keyframes santa-sac {
      0% { transform: rotate(5deg); }
      100% { transform: rotate(0deg); }
    }

    @keyframes particles {
      0% {
        transform: translate(20px, -20px);
        opacity: 1;
      }
      100% {
        transform: translate(-80px, 80px);
        opacity: 0;
      }
    }
  `;
  document.head.appendChild(sleighStyle);

  // Create the sleigh container
  const sleighContainer = document.createElement('div');
  sleighContainer.className = 'sleigh-container';
  sleighContainer.innerHTML = `
    <div class="sleigh-santa">
      <div class="santa santa--sleigh">
        <div class="santa__hat">
          <div class="santa__hat-part"></div>
          <div class="santa__hat-part"></div>
        </div>
        <div class="santa__face">
          <div class="santa__eyebrows santa__eyebrows--right"></div>
          <div class="santa__eyebrows santa__eyebrows--left"></div>
          <div class="santa__eye santa__eye--right"></div>
          <div class="santa__eye santa__eye--left"></div>
          <div class="santa__nose"></div>
          <div class="santa__cheek santa__cheek--right"></div>
          <div class="santa__cheek santa__cheek--left"></div>
          <div class="santa__beard">
            <div class="santa__beard-part"></div>
            <div class="santa__beard-part"></div>
            <div class="santa__beard-part"></div>
          </div>
          <div class="santa__mouth"></div>
        </div>
        <div class="santa__body">
          <div class="santa__body-top"></div>
          <div class="santa__hand santa__hand--left">
            <div class="santa__hand-inner"></div>
          </div>
          <div class="santa__hand santa__hand--right"></div>
        </div>
      </div>
      <div class="sleigh-feet"></div>
      <div class="lead">
        <div class="lead-inner"></div>
      </div>
      <div class="lead lead--back">
        <div class="lead-inner"></div>
      </div>  
      <div class="reindeer">
        <div class="reindeer__face">
          <div class="reindeer__ear"></div>
          <div class="reindeer__horn reindeer__horn--right"></div>
          <div class="reindeer__horn reindeer__horn--left"></div>
        </div>
        <div class="reindeer__body">
          <div class="reindeer__foot reindeer__foot--front">
            <div class="reindeer__Continue00:50foot-inner"></div>
</div>
<div class="reindeer__foot reindeer__foot--front reindeer__foot--inside">
<div class="reindeer__foot-inner"></div>
</div>
<div class="reindeer__foot reindeer__foot--back">
<div class="reindeer__foot-inner"></div>
</div>
<div class="reindeer__foot reindeer__foot--back reindeer__foot--inside">
<div class="reindeer__foot-inner"></div>
</div>
<div class="reindeer__tail"></div>
<div class="reindeer__spots"></div>
</div>
</div>
<div class="particles"></div>
</div>
`;
document.body.appendChild(sleighContainer);
// make the stars snowflakes moving downwards

const style = document.createElement('style');
    style.textContent = `
      @keyframes stars {
        0%   { transform: translateY(0px); }
        100% { transform: translateY(+2000px); }
      }
              @keyframes stars {
        0%   { transform: translateY(0); }
        100% { transform: translateY(3000px); }
      }

      #space-background::before {
        background-image: 
          radial-gradient(2px 2px at 17px 43px, #fff, transparent),
          radial-gradient(2px 2px at 39px 157px, #fff, transparent),
          radial-gradient(2px 2px at 73px 91px, #fff, transparent),
          radial-gradient(2px 2px at 109px 134px, #fff, transparent),
          radial-gradient(2px 2px at 143px 27px, #fff, transparent),
          radial-gradient(2px 2px at 53px 79px, #fff, transparent),
          radial-gradient(2px 2px at 85px 123px, #fff, transparent),
          radial-gradient(2px 2px at 174px 36px, #fff, transparent),
          radial-gradient(2px 2px at 102px 64px, #fff, transparent),
          radial-gradient(2px 2px at 127px 187px, #fff, transparent);
        background-size: 250px 250px;
        opacity: 0.6;
        animation: none !important; /* Static layer - no movement */
      }

      /* Moving larger/closer snowflakes layer (replaces Layer 2 stars) */
      #space-background::after {
        background-image: 
          radial-gradient(4px 4px at 43px 267px, #fff, transparent),
          radial-gradient(4px 4px at 167px 312px, #fff, transparent),
          radial-gradient(4px 4px at 213px 117px, #fff, transparent),
          radial-gradient(4px 4px at 81px 209px, #fff, transparent),
          radial-gradient(4px 4px at 239px 349px, #fff, transparent),
          radial-gradient(4px 4px at 124px 281px, #fff, transparent),
          radial-gradient(4px 4px at 198px 153px, #fff, transparent),
          radial-gradient(4px 4px at 56px 379px, #fff, transparent),
          radial-gradient(3px 3px at 297px 246px, #fff, transparent),
          radial-gradient(3px 3px at 147px 133px, #fff, transparent);
        background-size: 400px 400px;
        opacity: 0.8;
        animation: stars 180s linear infinite !important; /* Slower for gentle snow fall */
      }

    `;
    document.head.appendChild(style);


function addChristmasOrbitingTreats() {
  // Hide original orbits
  document.querySelectorAll('.orbit').forEach(orbit => {
    orbit.style.display = 'none';
  });

  const container = document.getElementById('planet-container');
  if (!container) return;

  // Helper to create a properly positioned orbiting image
  function createOrbitingImage(imgUrl, size, orbitRadius, speed, delay = 0) {
    const orbit = document.createElement('div');
    orbit.classList.add('orbit');
    orbit.style.width = `${orbitRadius * 2}px`;
    orbit.style.height = `${orbitRadius * 2}px`;
    orbit.style.position = 'absolute';
    orbit.style.left = '50%';
    orbit.style.top = '50%';
    orbit.style.transform = 'translate(-50%, -50%)';
    orbit.style.animation = `orbit ${speed}s linear infinite ${delay}s`;

    const img = document.createElement('img');
    img.src = imgUrl;
    img.style.width = `${size}px`;
    img.style.height = `${size}px`;
    img.style.objectFit = 'contain';
    img.style.position = 'absolute';
    img.style.left = '0';
    img.style.top = '50%';
    img.style.transform = 'translateY(-50%)';
    img.style.pointerEvents = 'none';
    img.style.filter = 'drop-shadow(0 0 20px rgba(255,255,255,0.7))';

    orbit.appendChild(img);
    container.appendChild(orbit);
  }

  // Better transparent PNGs for cleaner look
  createOrbitingImage('/static/xmas/c4.webp', 130, 1100, 40, -5);

  createOrbitingImage('/static/xmas/c3.webp', 100, 850, 55, -12);

  createOrbitingImage('/static/xmas/c2.webp', 120, 600, 65, -20);

  createOrbitingImage('/static/xmas/c1.webp', 110, 400, 32, -8);
}

function replaceBanner() {
  const banner = document.getElementById('banner');
  banner.src = '/static/xmas/stationeers-winter.webp';
}

addChristmasOrbitingTreats();
replaceBanner();

});