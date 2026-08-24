#ifndef _FCITX5_BAMBOO_VIET_ADDON_H_
#define _FCITX5_BAMBOO_VIET_ADDON_H_

#include <fcitx/inputmethodengine.h>
#include <fcitx/addonfactory.h>
#include <fcitx/addonmanager.h>
#include <fcitx/instance.h>
#include <fcitx/inputcontext.h>
#include <fcitx-utils/event.h>
#include <stdint.h>
#include <memory>
#include <string>

// Include C ABI from libvicore
#include "../bin/libvicore.h"

namespace fcitx {

class BambooVietState : public InputContextProperty {
public:
    BambooVietState(InputContext* ic, const std::string& inputMethod);
    ~BambooVietState();

    void reset();
    bool processKey(KeyEvent& keyEvent);

private:
    InputContext* ic_;
    uintptr_t engineHandle_;
    std::string inputMethod_;
};

class BambooVietEngine : public InputMethodEngineV3 {
public:
    BambooVietEngine(Instance* instance);
    ~BambooVietEngine() override;

    void keyEvent(const InputMethodEntry& entry, KeyEvent& keyEvent) override;
    void activate(const InputMethodEntry& entry, InputContextEvent& event) override;
    void deactivate(const InputMethodEntry& entry, InputContextEvent& event) override;
    void reset(const InputMethodEntry& entry, InputContextEvent& event) override;

    Instance* instance() const { return instance_; }

private:
    Instance* instance_;
    FactoryFor<BambooVietState> stateFactory_;
};

class BambooVietEngineFactory : public AddonFactory {
public:
    AddonInstance* create(AddonManager* manager) override {
        return new BambooVietEngine(manager->instance());
    }
};

} // namespace fcitx

#endif // _FCITX5_BAMBOO_VIET_ADDON_H_
