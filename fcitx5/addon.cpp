#include "addon.h"
#include <fcitx/inputcontext.h>
#include <fcitx/inputpanel.h>
#include <fcitx-utils/key.h>
#include <iostream>
#include <cstring>

namespace fcitx {

BambooVietState::BambooVietState(InputContext* ic, const std::string& inputMethod)
    : ic_(ic), engineHandle_(0), inputMethod_(inputMethod) {
    // Initialize libvicore engine with EstdFlags (7)
    engineHandle_ = vi_engine_new(const_cast<char*>(inputMethod_.c_str()), 7);
}

BambooVietState::~BambooVietState() {
    if (engineHandle_ != 0) {
        vi_engine_free(engineHandle_);
        engineHandle_ = 0;
    }
}

void BambooVietState::reset() {
    if (engineHandle_ != 0) {
        vi_engine_reset(engineHandle_);
    }
    if (ic_) {
        ic_->inputPanel().reset();
        ic_->updatePreedit();
        ic_->updateUserInterface(UserInterfaceComponent::InputPanel);
    }
}

bool BambooVietState::processKey(KeyEvent& keyEvent) {
    if (keyEvent.isRelease()) {
        return false;
    }

    Key key = keyEvent.key();
    unsigned int sym = key.sym();
    unsigned int state = key.states();

    char commitBuf[512] = {0};
    char preeditBuf[512] = {0};
    int bsCount = 0;

    int processed = vi_engine_process_key(
        engineHandle_,
        sym,
        key.code(),
        state,
        commitBuf, sizeof(commitBuf),
        preeditBuf, sizeof(preeditBuf),
        &bsCount
    );

    if (processed) {
        // If there is commit text (e.g. on word break or enter)
        if (strlen(commitBuf) > 0) {
            ic_->commitString(commitBuf);
        }

        // Update preedit string on the input panel
        Text preeditText;
        if (strlen(preeditBuf) > 0) {
            preeditText.append(preeditBuf, TextFormatFlag::Underline);
        }
        ic_->inputPanel().setClientPreedit(preeditText);
        ic_->updatePreedit();
        ic_->updateUserInterface(UserInterfaceComponent::InputPanel);

        keyEvent.filterAndAccept();
        return true;
    }

    return false;
}

BambooVietEngine::BambooVietEngine(Instance* instance)
    : instance_(instance),
      stateFactory_([this](InputContext& ic) {
          return new BambooVietState(&ic, "Telex");
      }) {
    instance_->inputContextManager().registerProperty("bambooVietState", &stateFactory_);
}

BambooVietEngine::~BambooVietEngine() = default;

void BambooVietEngine::keyEvent(const InputMethodEntry& entry, KeyEvent& keyEvent) {
    auto* ic = keyEvent.inputContext();
    if (!ic) {
        return;
    }

    auto* state = ic->propertyFor(&stateFactory_);
    if (state) {
        state->processKey(keyEvent);
    }
}

void BambooVietEngine::activate(const InputMethodEntry& entry, InputContextEvent& event) {
    auto* ic = event.inputContext();
    if (!ic) return;
    auto* state = ic->propertyFor(&stateFactory_);
    if (state) {
        state->reset();
    }
}

void BambooVietEngine::deactivate(const InputMethodEntry& entry, InputContextEvent& event) {
    auto* ic = event.inputContext();
    if (!ic) return;
    auto* state = ic->propertyFor(&stateFactory_);
    if (state) {
        state->reset();
    }
}

void BambooVietEngine::reset(const InputMethodEntry& entry, InputContextEvent& event) {
    auto* ic = event.inputContext();
    if (!ic) return;
    auto* state = ic->propertyFor(&stateFactory_);
    if (state) {
        state->reset();
    }
}

} // namespace fcitx

FCITX_ADDON_FACTORY(fcitx::BambooVietEngineFactory)
